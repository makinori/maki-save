package ffmpeg

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/ebitengine/purego"
)

// the *.so files come from
// https://central.sonatype.com/artifact/com.mrljdx/ffmpeg-kit-full/6.1.4
// https://repo1.maven.org/maven2/com/mrljdx/ffmpeg-kit-full/6.1.4/

var (
	// avformat
	avformat_open_input func(
		ps **AVFormatContext, url string, fmt *AVInputFormat, options **AVDictionary,
	) int32
	avformat_close_input          func(s **AVFormatContext)
	avformat_find_stream_info     func(ic *AVFormatContext, options **AVDictionary) int32
	av_dump_format                func(ic *AVFormatContext, index int32, url string, is_output int32)
	avcodec_find_decoder          func(id AVCodecID) *AVCodec
	avcodec_alloc_context3        func(codec *AVCodec) *AVCodecContext
	avcodec_free_context          func(avctx **AVCodecContext)
	avcodec_parameters_to_context func(codec *AVCodecContext, par *AVCodecParameters) int32
	avcodec_open2                 func(avctx *AVCodecContext, codec *AVCodec, options **AVDictionary) int32
	av_read_frame                 func(s *AVFormatContext, pkt *AVPacket) int32
	av_seek_frame                 func(s *AVFormatContext, stream_index int32, timestamp int64, flags int32) int32
	avio_alloc_context            func(
		buffer unsafe.Pointer, buffer_size int32, write_flag int32, opaque unsafe.Pointer,
		read_packet func(opaque, buffer unsafe.Pointer, buf_size int32) int32,
		write_packet func(opaque, buffer unsafe.Pointer, buf_size int32) int32,
		seek func(opaque unsafe.Pointer, offset int64, whence int) int64,
	) *AVIOContext
	avio_context_free      func(s **AVIOContext)
	avformat_alloc_context func() *AVFormatContext
	avformat_free_context  func(s *AVFormatContext)

	// avutil
	av_frame_alloc           func() *AVFrame
	av_frame_free            func(frame **AVFrame)
	av_image_get_buffer_size func(pix_fmt AVPixelFormat, width, height, align int32) int32
	av_image_fill_arrays     func(dst_data **uint8, dst_linesize *int32, src *uint8,
		pix_fmt AVPixelFormat, width, height, align int32) int32
	av_malloc func(size uintptr) unsafe.Pointer

	// swscale
	sws_getContext func(srcW, srcH int32, srcFormat AVPixelFormat,
		dstW, dstH int32, dstFormat AVPixelFormat,
		flags int32, srcFilter *SwsFilter,
		dstFilter *SwsFilter, param *float64) *SwsContext
	sws_freeContext func(swsContext *SwsContext)
	sws_scale       func(c *SwsContext, srcSlice **uint8,
		srcStride *int32, srcSliceY, srcSliceH int32,
		dst **uint8, dstStride *int32) int32

	// avcodec
	av_packet_alloc       func() *AVPacket
	av_packet_free        func(pkt **AVPacket)
	av_packet_unref       func(pkt *AVPacket)
	avcodec_send_packet   func(avctx *AVCodecContext, avpkt *AVPacket) int32
	avcodec_receive_frame func(avctx *AVCodecContext, frame *AVFrame) int32
	// avcodec_flush_buffers func(avctx *AVCodecContext)
)

var initialized = false

var (
	libsTmpDir string
	handles    = map[string]uintptr{}
)

func openLib(name string) (uintptr, error) {
	handle, ok := handles[name]
	if ok {
		return handle, nil
	}

	libPath := name

	if useEmbeddedLibraries {
		libPath = filepath.Join(libsTmpDir, name)
	}

	handle, err := libOpen(libPath)
	if err != nil {
		return 0, err
	}

	handles[name] = handle

	return handle, nil
}

func initFFmpeg() error {
	if initialized {
		return nil
	}
	initialized = true

	// make tmp dir
	// TODO: reuse same directory so we dont keep writing ffmpeg libs

	var err error

	if useEmbeddedLibraries {
		libsTmpDir, err = os.MkdirTemp("", "ffmpeg")
		if err != nil {
			closeFfmpeg()
			return err
		}
		fmt.Println("added tmp dir: " + libsTmpDir)

		libs, _ := fs.ReadDir(libsFS, ".")
		for _, lib := range libs {
			if lib.IsDir() {
				continue
			}
			data, _ := fs.ReadFile(libsFS, lib.Name())
			os.WriteFile(filepath.Join(libsTmpDir, lib.Name()), data, 0755)
		}
	}

	// order matters

	for _, name := range preloadLibs {
		_, err := openLib(name)
		if err != nil {
			closeFfmpeg()
			return err
		}
	}

	// bind

	r := purego.RegisterLibFunc

	avformat, err := openLib("libavformat.so")
	if err != nil {
		closeFfmpeg()
		return err
	}

	r(&avformat_open_input, avformat, "avformat_open_input")
	r(&avformat_close_input, avformat, "avformat_close_input")
	r(&avformat_find_stream_info, avformat, "avformat_find_stream_info")
	r(&av_dump_format, avformat, "av_dump_format")
	r(&avcodec_find_decoder, avformat, "avcodec_find_decoder")
	r(&avcodec_alloc_context3, avformat, "avcodec_alloc_context3")
	r(&avcodec_free_context, avformat, "avcodec_free_context")
	r(&avcodec_parameters_to_context, avformat, "avcodec_parameters_to_context")
	r(&avcodec_open2, avformat, "avcodec_open2")
	r(&av_read_frame, avformat, "av_read_frame")
	r(&av_seek_frame, avformat, "av_seek_frame")
	r(&avio_alloc_context, avformat, "avio_alloc_context")
	r(&avio_context_free, avformat, "avio_context_free")
	r(&avformat_alloc_context, avformat, "avformat_alloc_context")
	r(&avformat_free_context, avformat, "avformat_free_context")

	avutil, err := openLib("libavutil.so")
	if err != nil {
		closeFfmpeg()
		return err
	}

	r(&av_frame_alloc, avutil, "av_frame_alloc")
	r(&av_frame_free, avutil, "av_frame_free")
	r(&av_image_get_buffer_size, avutil, "av_image_get_buffer_size")
	r(&av_image_fill_arrays, avutil, "av_image_fill_arrays")
	r(&av_malloc, avutil, "av_malloc")

	swscale, err := openLib("libswscale.so")
	if err != nil {
		closeFfmpeg()
		return err
	}

	r(&sws_getContext, swscale, "sws_getContext")
	r(&sws_freeContext, swscale, "sws_freeContext")
	r(&sws_scale, swscale, "sws_scale")

	avcodec, err := openLib("libavcodec.so")
	if err != nil {
		closeFfmpeg()
		return err
	}

	r(&av_packet_alloc, avcodec, "av_packet_alloc")
	r(&av_packet_free, avcodec, "av_packet_free")
	r(&av_packet_unref, avcodec, "av_packet_unref")
	r(&avcodec_send_packet, avcodec, "avcodec_send_packet")
	r(&avcodec_receive_frame, avcodec, "avcodec_receive_frame")
	// r(&avcodec_flush_buffers, avcodec, "avcodec_flush_buffers")

	return nil
}

func closeFfmpeg() {
	if !initialized {
		return
	}

	for _, handle := range handles {
		libClose(handle)
	}

	handles = map[string]uintptr{}

	if libsTmpDir != "" {
		os.RemoveAll(libsTmpDir)
		fmt.Println("removed tmp dir: " + libsTmpDir)
		libsTmpDir = ""
	}

	initialized = false
}
