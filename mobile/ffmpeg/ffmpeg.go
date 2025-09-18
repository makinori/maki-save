package ffmpeg

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"unsafe"
)

func ffmpegMiddleFrame(filePath string) ([]byte, int, int, error) {
	err := initFFmpeg()
	if err != nil {
		return []byte{}, 0, 0, err
	}
	defer closeFfmpeg()

	var fmtCtx *AVFormatContext

	if avformat_open_input(&fmtCtx, filePath, nil, nil) != 0 {
		return []byte{}, 0, 0, errors.New("failed to open file")
	}
	defer avformat_close_input(&fmtCtx)

	if avformat_find_stream_info(fmtCtx, nil) < 0 {
		return []byte{}, 0, 0, errors.New("failed to find stream info")
	}

	// dump to stderr
	av_dump_format(fmtCtx, 0, filePath, 0)

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

	numBytes := av_image_get_buffer_size(
		AV_PIX_FMT_RGB24, codecCtx.width, codecCtx.height, 1,
	)

	var buffer = make([]byte, numBytes)
	av_image_fill_arrays(
		&frameRGB.data[0], &frameRGB.linesize[0], (*uint8)(&buffer[0]),
		AV_PIX_FMT_RGB24, codecCtx.width, codecCtx.height, 1,
	)

	swsCtx := sws_getContext(
		codecCtx.width, codecCtx.height, codecCtx.pix_fmt,
		codecCtx.width, codecCtx.height, AV_PIX_FMT_RGB24,
		SWS_BILINEAR, nil, nil, nil,
	)
	if swsCtx == nil {
		fmt.Println("failed to get sws context")
	}
	defer sws_freeContext(swsCtx)

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

		sws_scale(
			swsCtx, &frame.data[0], &frame.linesize[0], 0,
			codecCtx.height, &frameRGB.data[0], &frameRGB.linesize[0],
		)

		break
	}

	return buffer, int(codecCtx.width), int(codecCtx.height), nil
}

func GetMiddleFrameFromVideo(inputData []byte) ([]byte, error) {
	tmp, err := os.CreateTemp("", "")
	if err != nil {
		return []byte{}, err
	}
	tmp.Write(inputData)
	tmp.Close()
	fmt.Println("added tmp: " + tmp.Name())
	defer func() {
		os.Remove(tmp.Name())
		fmt.Println("removed tmp: " + tmp.Name())
	}()

	imageData, w, h, err := ffmpegMiddleFrame(tmp.Name())
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
