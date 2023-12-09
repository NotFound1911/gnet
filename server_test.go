package gnet

import (
	"errors"
	"github.com/NotFound1911/gnet/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

// $GOPATH/bin/mockgen -destination=mocks/net_conn.gen.go -package=mocks net Conn

func TestHandleConn(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(controller *gomock.Controller) net.Conn
		wantErr error
	}{
		{
			name: "read error",
			mock: func(controller *gomock.Controller) net.Conn {
				conn := mocks.NewMockConn(controller)
				conn.EXPECT().Read(gomock.Any()).Return(0, errors.New("read error"))
				return conn
			},
			wantErr: errors.New("read error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()
			err := handleConn(tc.mock(controller))
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
