package server

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/peer-calls/peer-calls/v4/server/logger"
	"github.com/peer-calls/peer-calls/v4/server/transport"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

func (p *WebRTCTransport) handleIncomingTrack(tr transport.TrackRemoteWithRTCPReader) {
	remoteTrack := tr.TrackRemote
	trackKind := remoteTrack.Track().Codec().MimeType
	packetChan := make(chan *rtp.Packet, 256*256)

	var writeFile func(chan *rtp.Packet)
	fmt.Println("trackKind:", trackKind)
	// Ghi video hoặc audio theo loại track
	if trackKind == webrtc.RTPCodecTypeAudio.String() || trackKind == "audio/opus" {
		writeFile = p.writeAudioToFile
	} else {
		//if trackKind == webrtc.RTPCodecTypeVideo.String() {
		writeFile = p.writeVideoToFile
	}

	// Goroutine để ghi file
	go writeFile(packetChan)

	// Goroutine để đọc RTP từ track và đẩy vào channel
	go func() {
		defer close(packetChan)
		for {
			pkt, _, err := remoteTrack.ReadRTP()
			if err != nil {
				if err != io.EOF {
					p.log.Trace("ReadRTP error", logger.Ctx{"error": err})
				}
				break
			}
			packetChan <- pkt
		}
	}()
}

func (p *WebRTCTransport) writeVideoToFile(packetChan chan *rtp.Packet) {
	fmt.Println("============ VÀO SAVE VIDEO ========", time.Now().Unix())

	// Mở file một lần duy nhất, không tạo lại mỗi lần có track mới
	filePath := "./recordings/session_video.ivf"
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		p.log.Trace("Error creating or opening video file", logger.Ctx{"error": err})
		return
	}
	defer file.Close()
	for pkt := range packetChan {
		raw, err := pkt.Marshal()
		if err != nil {
			p.log.Trace("Error marshaling RTP video", logger.Ctx{"error": err})
			continue
		}
		_, err = file.Write(raw)
		if err != nil {
			p.log.Trace("Error writing video file", logger.Ctx{"error": err})
			break
		}
	}
}

func (p *WebRTCTransport) writeAudioToFile(packetChan chan *rtp.Packet) {

	fmt.Println("============ VÀO SAVE AUDIO ========", time.Now().Unix())
	// Mở file một lần duy nhất, không tạo lại mỗi lần có track mới
	filePath := "./recordings/session_audio.ogg"
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		p.log.Trace("Error creating or opening audio file", logger.Ctx{"error": err})
		return
	}
	defer file.Close()
	for pkt := range packetChan {
		raw, err := pkt.Marshal()
		if err != nil {
			p.log.Trace("Error marshaling RTP audio", logger.Ctx{"error": err})
			continue
		}
		_, err = file.Write(raw)
		if err != nil {
			p.log.Trace("Error writing audio file", logger.Ctx{"error": err})
			break
		}
	}
}
