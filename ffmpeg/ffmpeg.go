package ffmpeg

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/png"
	"unsafe"
)

func goString(cString *uint8) string {
	var len uintptr
	for range 12 {
		if *(*uint8)(unsafe.Pointer(
			(uintptr)(unsafe.Pointer(cString)) + len,
		)) == 0 {
			break
		}
		len++
	}
	return string(bytes.Runes(unsafe.Slice(cString, len)))
}

func ffmpegMiddleFrame(fileData []byte) ([]byte, int, int, error) {
	err := initFFmpeg()
	if err != nil {
		return []byte{}, 0, 0, err
	}
	defer closeFfmpeg()

	ioPosition := 0
	ioBufferSize := 4096
	ioBuffer := av_malloc(uintptr(ioBufferSize))
	// dealloc? example doesnt free

	readPacket := func(_, bufferPtr unsafe.Pointer, bufferSize int32) int32 {
		buffer := unsafe.Slice((*byte)(bufferPtr), bufferSize)

		if ioPosition > len(fileData)-1 {
			return AVERROR_EOF
		}

		copied := copy(buffer, fileData[ioPosition:])
		ioPosition += copied

		if copied == 0 {
			return AVERROR_EOF
		}

		return int32(copied)
	}

	seek := func(_ unsafe.Pointer, offset int64, whence int) int64 {
		// maybe need to len()-1
		switch whence {
		case SEEK_SET:
			ioPosition = int(offset)
		case SEEK_CUR:
			ioPosition += int(offset)
		case SEEK_END:
			ioPosition = len(fileData) + int(offset)
		case AVSEEK_SIZE:
			return int64(len(fileData))
		default:
			fmt.Println("unknown seek whence", whence)
			return -1
		}
		ioPosition = min(ioPosition, len(fileData))
		return int64(ioPosition)
	}

	ioCtx := avio_alloc_context(
		ioBuffer,
		int32(ioBufferSize),
		0, // read only
		nil,
		readPacket,
		func(_, _ unsafe.Pointer, _ int32) int32 { return 0 },
		seek,
	)
	if ioCtx == nil {
		return []byte{}, 0, 0, errors.New("failed to allocate io context")
	}
	defer avio_context_free(&ioCtx)

	fmtCtx := avformat_alloc_context()
	if fmtCtx == nil {
		return []byte{}, 0, 0, errors.New("failed to allocate format context")
	}
	// defer avformat_free_context(fmtCtx) // causes crash. example doesnt free

	fmtCtx.pb = ioCtx

	if avformat_open_input(&fmtCtx, "", nil, nil) != 0 {
		return []byte{}, 0, 0, errors.New("failed to open file")
	}
	defer avformat_close_input(&fmtCtx)

	if avformat_find_stream_info(fmtCtx, nil) < 0 {
		return []byte{}, 0, 0, errors.New("failed to find stream info")
	}

	// dump to stderr
	av_dump_format(fmtCtx, 0, "", 0)

	streams := unsafe.Slice(fmtCtx.streams, fmtCtx.nb_streams)

	var videoStreamIndex int32 = -1
	for i, stream := range streams {
		if stream.codecpar.codec_type == AVMEDIA_TYPE_VIDEO {
			videoStreamIndex = int32(i)
			break
		}
	}
	if videoStreamIndex == -1 {
		return []byte{}, 0, 0, errors.New("failed to find video stream")
	}

	videoStream := streams[videoStreamIndex]

	// we want the middle frame
	frameTimestamp := videoStream.duration / 2

	// seek will jump the nearest keyframe before this frame
	// so we still need to decode a few frames until we get to our timestamp
	// TODO: this is actually kinda slow and i dont think file managers do this

	if av_seek_frame(
		fmtCtx, videoStreamIndex, frameTimestamp, AVSEEK_FLAG_BACKWARD,
	) < 0 {
		return []byte{}, 0, 0, errors.New("failed to seek to middle frame")
	}

	codecParams := videoStream.codecpar

	codec := avcodec_find_decoder(codecParams.codec_id)
	if codec == nil {
		return []byte{}, 0, 0, errors.New("failed to find decoder")
	}

	codecCtx := avcodec_alloc_context3(codec)
	if codecCtx == nil {
		return []byte{}, 0, 0, errors.New("failed to allocate codec context")
	}
	defer avcodec_free_context(&codecCtx)

	if avcodec_parameters_to_context(codecCtx, codecParams) < 0 {
		return []byte{}, 0, 0, errors.New("failed to copy codec params to context")
	}

	if avcodec_open2(codecCtx, codec, nil) < 0 {
		return []byte{}, 0, 0, errors.New("failed to open codec")
	}

	frame := av_frame_alloc()
	frameRGB := av_frame_alloc()
	if frame == nil || frameRGB == nil {
		return []byte{}, 0, 0, errors.New("failed to allocate frame")
	}
	defer av_frame_free(&frame)
	defer av_frame_free(&frameRGB)

	packet := av_packet_alloc()
	if packet == nil {
		fmt.Println("failed to allocate packet")
	}
	defer av_packet_free(&packet)

	for av_read_frame(fmtCtx, packet) >= 0 {
		if packet.stream_index != videoStreamIndex {
			av_packet_unref(packet)
			continue
		}

		ret := avcodec_send_packet(codecCtx, packet)
		if ret == -EAGAIN {
			continue
		} else if ret < 0 {
			return []byte{}, 0, 0, errors.New("failed to send packet to decoder")
		}

		ret = avcodec_receive_frame(codecCtx, frame)
		if ret == -EAGAIN {
			continue
		} else if ret != 0 {
			return []byte{}, 0, 0, errors.New("failed to receive frame")
		}

		if frame.key_frame < frameTimestamp {
			continue
		}

		// got frame we want to use

		break
	}

	// pix_fmt can be -1 if we do this before decoding at least one frame
	swsCtx := sws_getContext(
		frame.width, frame.height, codecCtx.pix_fmt,
		frame.width, frame.height, AV_PIX_FMT_RGB24,
		SWS_BILINEAR, nil, nil, nil,
	)
	if swsCtx == nil {
		fmt.Println("failed to get sws context")
	}
	defer sws_freeContext(swsCtx)

	numBytes := av_image_get_buffer_size(
		AV_PIX_FMT_RGB24, frame.width, frame.height, 1,
	)

	var buffer = make([]byte, numBytes)
	av_image_fill_arrays(
		&frameRGB.data[0], &frameRGB.linesize[0], (*uint8)(&buffer[0]),
		AV_PIX_FMT_RGB24, frame.width, frame.height, 1,
	)

	sws_scale(
		swsCtx, &frame.data[0], &frame.linesize[0], 0,
		frame.height, &frameRGB.data[0], &frameRGB.linesize[0],
	)

	return buffer, int(frame.width), int(frame.height), nil
}

func GetMiddleFrameFromVideo(inputData []byte) ([]byte, error) {
	imageData, w, h, err := ffmpegMiddleFrame(inputData)
	if err != nil {
		return []byte{}, err
	}

	// return []byte{}, nil

	image := image.NewRGBA(
		image.Rect(0, 0, w, h),
	)

	for i := range len(imageData) / 3 {
		image.Pix[i*4] = imageData[i*3]
		image.Pix[i*4+1] = imageData[i*3+1]
		image.Pix[i*4+2] = imageData[i*3+2]
		image.Pix[i*4+3] = 255
	}

	pngData := bytes.NewBuffer(nil)
	err = png.Encode(pngData, image)
	if err != nil {
		return []byte{}, errors.New("failed to encode to png")
	}

	return pngData.Bytes(), nil
}
