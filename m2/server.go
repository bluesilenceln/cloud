package m2

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func Health(writer http.ResponseWriter, request *http.Request) {
	for k, vs := range request.Header {
		for _, v := range vs {
			writer.Header().Set(k, v)
		}
	}

	version := os.Getenv("VERSION")
	writer.Header().Set("Version", version)

	ip := GetClientIP(request)
	statusCode := http.StatusOK
	writer.WriteHeader(statusCode)
	writer.Write([]byte("200"))
	fmt.Printf("client: %s, response status: %d\n", ip, statusCode)
}

type Server struct {
	*http.Server

	addr   string
	router *mux.Router
	ctx    context.Context
	cancel context.CancelFunc
}

type Handler struct {
	Path        string
	HandlerFunc http.HandlerFunc
}

func NewServer(addr string) *Server {
	router := mux.NewRouter()
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		addr:   addr,
		ctx:    ctx,
		cancel: cancel,
		router: router,
	}
}

func (s *Server) Register(handlers ...Handler) {
	for _, h := range handlers {
		s.router.HandleFunc(h.Path, h.HandlerFunc)
	}
}

func (s *Server) Run() {
	s.Server = &http.Server{
		Addr:    s.addr,
		Handler: s.router,
	}

	go func() {
		defer s.cancel()
		_ = s.Server.ListenAndServe()
		fmt.Println("server is stop")
	}()

	signC := make(chan os.Signal)
	signal.Notify(signC, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case sig := <-signC:
		fmt.Println("server recv signal:", sig)
		return
	case <-s.ctx.Done():
		fmt.Println("server context is canceled")
	}
}

func (s *Server) Stop() {
	s.Server.Shutdown(context.Background())
}

func GetClientIP(req *http.Request) string {
	xff := req.Header.Get("X-Forwarded-For")
	clientIP := strings.TrimSpace(strings.Split(xff, ",")[0])
	if clientIP != "" {
		return clientIP
	}

	clientIP = strings.TrimSpace(req.Header.Get("X-Real-Ip"))
	if clientIP != "" {
		return clientIP
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}
