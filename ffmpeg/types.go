package ffmpeg

import "unsafe"

/*
// #cgo CFLAGS: -I/usr/include/
// #cgo LDFLAGS: -lavcodec -lavformat -lswscale -lavutil
// #include <libavcodec/avcodec.h>
// #include <libavformat/avformat.h>
// #include <libswscale/swscale.h>
// #include <libavutil/imgutils.h>
import "C"
*/

const (
	AVMEDIA_TYPE_VIDEO = 0x0

	AV_PIX_FMT_RGB24 = 2

	SWS_FAST_BILINEAR = 1
	SWS_BILINEAR      = 2

	EAGAIN = 11

	AVSEEK_FLAG_BACKWARD = 0x1

	// AV_PKT_FLAG_KEY = 0x1

	SEEK_SET    = 0
	SEEK_CUR    = 1
	SEEK_END    = 2
	AVSEEK_SIZE = 0x10000
)

func MKTAG(a, b, c, d rune) int32 {
	return (int32)(a) | ((int32)(b) << 8) | ((int32)(c) << 16) | ((int32)(d) << 24)
}

var (
	// AV_TIME_BASE_Q = AVRational{
	// 	num: 1,
	// 	den: 1000000,
	// }

	AVERROR_EOF = -MKTAG('E', 'O', 'F', ' ')
)

func av_q2d(rational AVRational) float64 {
	return float64(rational.num) / float64(rational.den)
}

// type AVPacketSideDataType uint32

// const (
// 	AV_PKT_DATA_PALETTE AVPacketSideDataType = iota
// 	AV_PKT_DATA_NEW_EXTRADATA
// 	AV_PKT_DATA_PARAM_CHANGE
// 	AV_PKT_DATA_H263_MB_INFO
// 	AV_PKT_DATA_REPLAYGAIN
// 	AV_PKT_DATA_DISPLAYMATRIX
// 	// there are more
// )

const (
	AV_DICT_MATCH_CASE      = 1
	AV_DICT_IGNORE_SUFFIX   = 2
	AV_DICT_DONT_STRDUP_KEY = 4
	AV_DICT_DONT_STRDUP_VAL = 8
	AV_DICT_DONT_OVERWRITE  = 16
	AV_DICT_APPEND          = 32
	AV_DICT_MULTIKEY        = 64
)

type AVBufferRef struct{}
type AVCodec struct{}
type AVCodecInternal struct{}
type AVChapter struct{}
type AVClass struct{}
type AVDictionary struct{}
type AVInputFormat struct{}
type AVIOContext struct{}
type AVIOInterruptCB struct{}
type AVOutputFormat struct{}
type AVPacketSideData struct{}
type AVProgram struct{}
type AVStreamGroup struct{}

type SwsFilter struct{}
type SwsContext struct{}

type AVCodecID uint32
type AVPixelFormat int32

// these are not fully accurate, but work

type AVCodecContext struct {
	av_class            *AVClass
	log_level_offset    int32
	codec_type          int32
	codec               *AVCodec
	codec_id            uint32
	codec_tag           uint32
	priv_data           unsafe.Pointer
	internal            *AVCodecInternal
	opaque              unsafe.Pointer
	bit_rate            int64
	flags               int32
	flags2              int32
	extradata           *uint8
	extradata_size      int32
	time_base           AVRational
	pkt_timebase        AVRational
	framerate           AVRational
	ticks_per_frame     int32
	delay               int32
	width               int32
	height              int32
	coded_width         int32
	coded_height        int32
	sample_aspect_ratio AVRational
	pix_fmt             AVPixelFormat // int32
	// sw_pix_fmt                  int32
	// color_primaries             uint32
	// color_trc                   uint32
	// colorspace                  uint32
	// color_range                 uint32
	// chroma_sample_location      uint32
	// field_order                 uint32
	// refs                        int
	// has_b_frames                int
	// slice_flags                 int
	// draw_horiz_band             *[0]byte
	// get_format                  *[0]byte
	// max_b_frames                int
	// b_quant_factor              float
	// b_quant_offset              float
	// i_quant_factor              float
	// i_quant_offset              float
	// lumi_masking                float
	// temporal_cplx_masking       float
	// spatial_cplx_masking        float
	// p_masking                   float
	// dark_masking                float
	// nsse_weight                 int
	// me_cmp                      int
	// me_sub_cmp                  int
	// mb_cmp                      int
	// ildct_cmp                   int
	// dia_size                    int
	// last_predictor_count        int
	// me_pre_cmp                  int
	// pre_dia_size                int
	// me_subpel_quality           int
	// me_range                    int
	// mb_decision                 int
	// intra_matrix                *uint16_t
	// inter_matrix                *uint16_t
	// chroma_intra_matrix         *uint16_t
	// intra_dc_precision          int
	// mb_lmin                     int
	// mb_lmax                     int
	// bidir_refine                int
	// keyint_min                  int
	// gop_size                    int
	// mv0_threshold               int
	// slices                      int
	// sample_rate                 int
	// sample_fmt                  int32
	// ch_layout                   AVChannelLayout
	// frame_size                  int
	// block_align                 int
	// cutoff                      int
	// audio_service_type          uint32
	// request_sample_fmt          int32
	// initial_padding             int
	// trailing_padding            int
	// seek_preroll                int
	// get_buffer2                 *[0]byte
	// bit_rate_tolerance          int
	// global_quality              int
	// compression_level           int
	// qcompress                   float
	// qblur                       float
	// qmin                        int
	// qmax                        int
	// max_qdiff                   int
	// rc_buffer_size              int
	// rc_override_count           int
	// rc_override                 *RcOverride
	// rc_max_rate                 int64_t
	// rc_min_rate                 int64_t
	// rc_max_available_vbv_use    float
	// rc_min_vbv_overflow_use     float
	// rc_initial_buffer_occupancy int
	// trellis                     int
	// stats_out                   *char
	// stats_in                    *char
	// workaround_bugs             int
	// strict_std_compliance       int
	// error_concealment           int
	// debug                       int
	// err_recognition             int
	// hwaccel                     *AVHWAccel
	// hwaccel_context             unsafe.Pointer
	// hw_frames_ctx               *AVBufferRef
	// hw_device_ctx               *AVBufferRef
	// hwaccel_flags               int
	// extra_hw_frames             int
	// error                       [8]uint64_t
	// dct_algo                    int
	// idct_algo                   int
	// bits_per_coded_sample       int
	// bits_per_raw_sample         int
	// thread_count                int
	// thread_type                 int
	// active_thread_type          int
	// execute                     *[0]byte
	// execute2                    *[0]byte
	// profile                     int
	// level                       int
	// properties                  uint
	// skip_loop_filter            int32
	// skip_idct                   int32
	// skip_frame                  int32
	// skip_alpha                  int
	// skip_top                    int
	// skip_bottom                 int
	// lowres                      int
	// codec_descriptor            *AVCodecDescriptor
	// sub_charenc                 *char
	// sub_charenc_mode            int
	// subtitle_header_size        int
	// subtitle_header             *uint8_t
	// dump_separator              *uint8_t
	// codec_whitelist             *char
	// coded_side_data             *AVPacketSideData
	// nb_coded_side_data          int
	// export_side_data            int
	// max_pixels                  int64_t
	// apply_cropping              int
	// discard_damaged_percentage  int
	// max_samples                 int64_t
	// get_encode_buffer           *[0]byte
	// frame_num                   int64_t
	// side_data_prefer_packet     *int
	// nb_side_data_prefer_packet  uint
	// decoded_side_data           **AVFrameSideData
	// nb_decoded_side_data        int
	// _                           [4]byte
}

type AVCodecParameters struct {
	codec_type int32
	codec_id   AVCodecID
	codec_tag  uint32
	// extradata             *uint8_t
	// extradata_size        int
	// coded_side_data       *AVPacketSideData
	// nb_coded_side_data    int
	// format                int
	// bit_rate              int64_t
	// bits_per_coded_sample int
	// bits_per_raw_sample   int
	// profile               int
	// level                 int
	// width                 int
	// height                int
	// sample_aspect_ratio   AVRational
	// framerate             AVRational
	// field_order           uint32
	// color_range           uint32
	// color_primaries       uint32
	// color_trc             uint32
	// color_space           uint32
	// chroma_location       uint32
	// video_delay           int
	// ch_layout             AVChannelLayout
	// sample_rate           int
	// block_align           int
	// frame_size            int
	// initial_padding       int
	// trailing_padding      int
	// seek_preroll          int
}

type AVFormatContext struct {
	av_class   *AVClass
	iformat    *AVInputFormat
	oformat    *AVOutputFormat
	priv_data  unsafe.Pointer
	pb         *AVIOContext
	ctx_flags  int32
	nb_streams uint16
	streams    **AVStream
	// nb_stream_groups                uint
	// stream_groups                   **AVStreamGroup
	// nb_chapters                     uint
	// chapters                        **AVChapter
	// url                             *byte
	// startime                        int64
	// duration                        int64
	// bit_rate                        int64
	// packet_size                     uint
	// max_delay                       int
	// flags                           int
	// probesize                       int64
	// max_analyze_duration            int64
	// key                             *uint8
	// keylen                          int
	// nb_programs                     uint
	// programs                        **AVProgram
	// video_codec_id                  uint32
	// audio_codec_id                  uint32
	// subtitle_codec_id               uint32
	// data_codec_id                   uint32
	// metadata                        *AVDictionary
	// startime_realtime               int64
	// fps_probe_size                  int
	// error_recognition               int
	// interrupt_callback              AVIOInterruptCB
	// debug                           int
	// max_streams                     int
	// max_index_size                  uint
	// max_picture_buffer              uint
	// max_interleave_delta            int64
	// maxs_probe                      int
	// max_chunk_duration              int
	// max_chunk_size                  int
	// max_probe_packets               int
	// strict_std_compliance           int
	// event_flags                     int
	// avoid_negatives                 int
	// audio_preload                   int
	// use_wallclock_asimestamps       int
	// skip_estimate_duration_from_pts int
	// avio_flags                      int
	// duration_estimation_method      uint32
	// skip_initial_bytes              int64
	// corrects_overflow               uint
	// seek2any                        int
	// flush_packets                   int
	// probe_score                     int
	// format_probesize                int
	// codec_whitelist                 *byte
	// format_whitelist                *byte
	// protocol_whitelist              *byte
	// protocol_blacklist              *byte
	// io_repositioned                 int
	// video_codec                     *AVCodec
	// audio_codec                     *AVCodec
	// subtitle_codec                  *AVCodec
	// data_codec                      *AVCodec
	// metadata_header_padding         int
	// opaque                          unsafe.Pointer
	// control_message_cb              *byte
	// outputs_offset                  int64
	// dump_separator                  *uint8
	// io_open                         *byte
	// io_close2                       *byte
	// duration_probesize              int64
}

type AVFrame struct {
	data          [8]*uint8
	linesize      [8]int32
	extended_data **uint8
	width         int32
	height        int32
	nb_samples    int32
	format        int32
	key_frame     int64
	// pict_type             uint32
	// sample_aspect_ratio   AVRational
	// pts                   int64_t
	// pkt_dts               int64_t
	// time_base             AVRational
	// quality               int
	// opaque                unsafe.Pointer
	// repeat_pict           int
	// interlaced_frame      int
	// top_field_first       int
	// palette_has_changed   int
	// sample_rate           int
	// buf                   [8]*AVBufferRef
	// extended_buf          **AVBufferRef
	// nb_extended_buf       int
	// side_data             **AVFrameSideData
	// nb_side_data          int
	// flags                 int
	// color_range           uint32
	// color_primaries       uint32
	// color_trc             uint32
	// colorspace            uint32
	// chroma_location       uint32
	// best_effort_timestamp int64_t
	// pkt_pos               int64_t
	// metadata              *AVDictionary
	// decode_error_flags    int
	// pkt_size              int
	// hw_frames_ctx         *AVBufferRef
	// opaque_ref            *AVBufferRef
	// crop_top              size_t
	// crop_bottom           size_t
	// crop_left             size_t
	// crop_right            size_t
	// private_ref           *AVBufferRef
	// ch_layout             AVChannelLayout
	// duration              int64_t
}

type AVPacket struct {
	buf          *AVBufferRef
	pts          int64
	dts          int64
	data         *uint8
	size         int32
	stream_index int32
	// flags           int
	// side_data       *AVPacketSideData
	// side_data_elems int
	// duration        int64
	// pos             int64
	// opaque          unsafe.Pointer
	// opaque_ref      *AVBufferRef
	// time_base       AVRational
}

type AVRational struct {
	num int32
	den int32
}

type AVStream struct {
	av_class            *AVClass
	index               int32
	id                  int32
	codecpar            *AVCodecParameters
	priv_data           unsafe.Pointer
	time_base           AVRational
	start_time          int64
	duration            int64
	nb_frames           int64
	disposition         int32 // int
	discard             int32
	sample_aspect_ratio AVRational
	metadata            *AVDictionary
	// avg_frame_rate      AVRational
	// attached_pic        AVPacket
	// side_data           *AVPacketSideData
	// nb_side_data        int
	// event_flags         int
	// r_frame_rate        AVRational
	// pts_wrap_bits       int
	// _                   [4]byte
}

type AVDictionaryEntry struct {
	key  *uint8
	char *uint8
}
