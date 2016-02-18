package axihomeserver

import (
	"strings"

	"github.com/think-free/axihome/server/axihomehandler"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/server"
	"github.com/think-free/log"
)

const (
	hostname = "axihome"
)

type AxihomeServer struct {

	/* Map for routing by name */

	channelByDst map[string]chan jsonrpcmessage.RoutedMessage

	/* App handler */

	appHandler        *axihomehandler.Handler // Handler for messages destinated to axihome
	appHandlerChannel chan jsonrpcmessage.RoutedMessage

	/* Core server's channels */

	coreChannels2m      chan jsonrpcmessage.RoutedMessage // Channel for communication core to main
	coreChannelm2s      chan jsonrpcmessage.RoutedMessage // Channel for communication main to core
	coreInternalChannel chan jsonrpcmessage.RoutedMessage // Channel for communication core to core-handler
	coreClientState     chan jsonrpcserver.ClientState    // Channel for client state

	/* Backend server's channels */

	backendChannels2m      chan jsonrpcmessage.RoutedMessage // Channel for communication backend to main
	backendChannelm2s      chan jsonrpcmessage.RoutedMessage // Channel for communication main to backend
	backendInternalChannel chan jsonrpcmessage.RoutedMessage // Channel for communication backend to backend-handler
	backendClientState     chan jsonrpcserver.ClientState    // Channel for client state

	/* Frontend server's channels */

	frontendChannels2m      chan jsonrpcmessage.RoutedMessage // Channel for communication frontend to main
	frontendChannelm2s      chan jsonrpcmessage.RoutedMessage // Channel for communication main to frontend
	frontendInternalChannel chan jsonrpcmessage.RoutedMessage // Channel for communication frontend to frontend-handler
	frontendClientState     chan jsonrpcserver.ClientState    // Channel for client state
}

func New() *AxihomeServer {

	// Creation of server struct

	var server *AxihomeServer = &AxihomeServer{

		channelByDst: make(map[string]chan jsonrpcmessage.RoutedMessage),

		appHandlerChannel: make(chan jsonrpcmessage.RoutedMessage),

		coreChannels2m:      make(chan jsonrpcmessage.RoutedMessage),
		coreChannelm2s:      make(chan jsonrpcmessage.RoutedMessage),
		coreInternalChannel: make(chan jsonrpcmessage.RoutedMessage),
		coreClientState:     make(chan jsonrpcserver.ClientState),

		frontendChannels2m:      make(chan jsonrpcmessage.RoutedMessage),
		frontendChannelm2s:      make(chan jsonrpcmessage.RoutedMessage),
		frontendInternalChannel: make(chan jsonrpcmessage.RoutedMessage),
		frontendClientState:     make(chan jsonrpcserver.ClientState),

		backendChannels2m:      make(chan jsonrpcmessage.RoutedMessage),
		backendChannelm2s:      make(chan jsonrpcmessage.RoutedMessage),
		backendInternalChannel: make(chan jsonrpcmessage.RoutedMessage),
		backendClientState:     make(chan jsonrpcserver.ClientState),
	}

	server.appHandler = axihomehandler.New(&server.appHandlerChannel)

	// Maps of channels by destination to handle routing

	server.channelByDst["core"] = server.coreChannelm2s
	server.channelByDst["backend"] = server.backendChannelm2s
	server.channelByDst["frontend"] = server.frontendChannelm2s

	return server
}

func (server *AxihomeServer) Run() {

	var state = &jsonrpcmessage.StateBody{
		Domain: hostname,
		Tld:    true,
		Logged: true,
		Ssid:   "",
	}

	// Starting servers

	core := jsonrpcserver.New("0.0.0.0", "3330", "3331", "core", *state, server.coreChannels2m, server.coreChannelm2s, nil, nil, server.coreClientState)
	go core.Run()

	backend := jsonrpcserver.New("0.0.0.0", "3332", "3333", "backend", *state, server.backendChannels2m, server.backendChannelm2s, nil, nil, server.backendClientState)
	go backend.Run()

	frontend := jsonrpcserver.New("0.0.0.0", "3334", "3335", "frontend", *state, server.frontendChannels2m, server.frontendChannelm2s, nil, nil, server.frontendClientState)
	go frontend.Run()

	// Handling channels messages

	for {

		select {

		// Routing messages
		case mes := <-server.coreChannels2m:
			//log.Println("Message received from coreChannel :", mes.Type, " dst : ", mes.Dst)
			go server.routeMessage(&mes)
		case mes := <-server.backendChannels2m:
			//log.Println("Message received from backendChannel :", mes.Type, " dst : ", mes.Dst)
			go server.routeMessage(&mes)
		case mes := <-server.frontendChannels2m:
			//log.Println("Message received from frontendChannel :", mes.Type, " dst : ", mes.Dst)
			go server.routeMessage(&mes)

		// Receiving messages from internal handlers
		case mes := <-server.appHandlerChannel:
			go server.routeMessage(&mes)

		// Client state -> new client connected or client disconnected
		case state := <-server.coreClientState:
			log.Println("Client in core :", state.Name, "connected :", state.Connected)
		case state := <-server.backendClientState:
			log.Println("Client in backend :", state.Name, "connected :", state.Connected)
		case state := <-server.frontendClientState:
			log.Println("Client in frontend :", state.Name, "connected :", state.Connected)
		}
	}
}

func (server *AxihomeServer) routeMessage(mes *jsonrpcmessage.RoutedMessage) {

	if mes.Dst == hostname {

		if mes.Type == "rpc" {

			server.appHandler.Rpc(*mes)
		} else {

			log.Println("Unknow message type received for application : " + mes.Type)
		}

	} else {

		// Getting route for the message and sending it to the appropriated channel
		dst := strings.TrimSuffix(mes.Dst, "."+hostname)
		dstslice := strings.Split(dst, ".")
		dst = dstslice[len(dstslice)-1]

		channel := server.channelByDst[dstslice[len(dstslice)-1]]
		if channel != nil {

			channel <- *mes
		} else {

			log.Println("Channel not found for " + dst)
		}
	}
}
