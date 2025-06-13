// Copyright 2025 TII (SSRC) and the Ghaf contributors
// SPDX-License-Identifier: Apache-2.0

package evdevproxy

import (
	"context"
	"net"
	"time"
	//"os"

	givc_evdev "givc/modules/api/evdev"
	//givc_grpc "givc/modules/pkgs/grpc"

	//"golang.org/x/net/context"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	//"github.com/holoplot/go-evdev"
	"github.com/bendahl/uinput"
	"github.com/gvalkov/golang-evdev"
)

type evdevServiceClient struct {
	cc grpc.ClientConnInterface
}

type JoystickServer struct {
	givc_evdev.UnimplementedJoystickServiceServer
}

func (s *JoystickServer) StreamEvents(stream givc_evdev.JoystickService_StreamEventsServer) error {
	for {
		event, err := stream.Recv()

		if err != nil {
			return err
		}

		log.Infof("vunny Received event: %+v\n", event)

		// switch event.Code {
		// case 0: // ABS_X
		// 	s.dev.MoveX(int(event.Value))
		// case 1: // ABS_Y
		// 	s.dev.MoveY(int(event.Value))
		// case 304: // BTN_A (BTN_SOUTH)
		// 	s.dev.BtnSouthPress()
		// 	s.dev.BtnSouthRelease()
		// default:
		// 	log.Errorf("Unknown code: %d\n", event.Code)
		// }
	}
}

func NewEvdevProxyServer(runAsServer bool) {

	// Create a new evdev proxy controller
	NewEvdevProxyController(runAsServer)

	if !runAsServer {
		log.Infof("This code is for producer/client and should run in audio-vm only")
		conn, err := grpc.Dial("192.168.100.101:9191", grpc.WithInsecure())
		if err != nil {
			log.Errorf("vunny failed to dial to docker-vm:", err)
		} else {
			log.Infof("vunny connection to docker-vm was successful")
		}
		defer conn.Close()

		client := givc_evdev.NewJoystickServiceClient(conn)
		log.Infof("vunny got client", client)
		stream, err := client.StreamEvents(context.Background())
		if err != nil {
			log.Errorf("vunny Failed to open stream:", err)
		}

		log.Infof("vunny now trying to list input devices!")

		// devices, _ := evdev.ListInputDevices()

		// var dev *evdev.InputDevice
		// for _, d := range devices {
		// 	for ctype, _ := range d.Capabilities {
		// 		log.Infof("vunny ctype for device:", ctype.Type, ctype.Name)
		// 		if ctype.Type == evdev.EV_ABS {
		// 			dev = d
		// 			log.Infof("vunny device: %s supports ABS events:\n", d.Name)
		// 			break
		// 		}
		// 	}
		// }
		dev, err := evdev.Open("/dev/input/event6")
		if err != nil {
			log.Fatalf("vunny failed to open input device: %v", err)
		}
		//defer dev.Close()

		if dev == nil {
			log.Infof("vunny no joystick device found!")
		}
		log.Infof("vunny device allocated %+v", dev)

		for {
			events, err := dev.Read()
			if err != nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			for _, e := range events {
				log.Infof("vunny now sending input events over grpc", e.Type, e.Code, e.Value)
				if e.Type == evdev.EV_ABS || e.Type == evdev.EV_KEY {
					stream.Send(&givc_evdev.JoystickEvent{
						Type:  int32(e.Type),
						Code:  int32(e.Code),
						Value: int32(e.Value),
					})
				}
			}
		}
	} else {
		log.Infof("vunny inside client code")
		dev, err := uinput.CreateGamepad("/dev/uinput", []byte("Virtual Gamepad"), 3034, 33107)
		if err != nil {
			log.Fatalf("vunny error in creating device:", err)
		}
		defer dev.Close()

		grpcServer := grpc.NewServer()
		givc_evdev.RegisterJoystickServiceServer(grpcServer, &JoystickServer{})

		lis, err := net.Listen("tcp", ":9191")
		if err != nil {
			log.Fatalf("vunny Failed to listen:", err)
		}
		log.Infof("vunny gRPC server listening on port 9191")
		grpcServer.Serve(lis)
	}

}
