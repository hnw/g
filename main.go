// Package main implements a Google Assistant gRPC client server.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	pb "google.golang.org/genproto/googleapis/assistant/embedded/v1alpha2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/keepalive"
)

// Config 構造体（環境変数用にフラット化）
type Config struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
	DeviceID     string
	ModelID      string
	Language     string
}

type AssistantServer struct {
	client pb.EmbeddedAssistantClient
	config Config
}

func main() {
	// 1. 環境変数からの設定読み込み
	config := Config{
		ClientID:     os.Getenv("GAPROXY_CLIENT_ID"),
		ClientSecret: os.Getenv("GAPROXY_CLIENT_SECRET"),
		RefreshToken: os.Getenv("GAPROXY_REFRESH_TOKEN"),
		DeviceID:     getEnv("GAPROXY_DEVICE_ID", "default"),
		ModelID:      getEnv("GAPROXY_DEVICE_MODEL_ID", "default"),
		Language:     getEnv("GAPROXY_LANGUAGE_CODE", "en-US"),
	}

	// 必須項目のチェック
	if config.ClientID == "" || config.ClientSecret == "" || config.RefreshToken == "" {
		log.Fatal(
			"Missing required environment variables: GAPROXY_CLIENT_ID, GAPROXY_CLIENT_SECRET, GAPROXY_REFRESH_TOKEN",
		)
	}
	// 2. OAuth設定
	oauthConf := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
		Scopes: []string{"https://www.googleapis.com/auth/assistant-sdk-prototype"},
	}
	tokenSource := oauthConf.TokenSource(context.Background(), &oauth2.Token{
		RefreshToken: config.RefreshToken,
	})

	kacp := keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             time.Second,
		PermitWithoutStream: true,
	}

	// 3. gRPC接続
	conn, err := grpc.NewClient(
		"embeddedassistant.googleapis.com:443",
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
		grpc.WithPerRPCCredentials(oauth.TokenSource{TokenSource: tokenSource}),
		grpc.WithKeepaliveParams(kacp),
	)
	if err != nil {
		log.Fatalf("Failed to dial gRPC: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close gRPC connection: %v", err)
		}
	}()

	server := &AssistantServer{
		client: pb.NewEmbeddedAssistantClient(conn),
		config: config,
	}

	// 4. HTTPハンドラー
	http.HandleFunc("/", server.handleRoot)

	port := getEnv("PORT", "8080")
	log.Printf("Server listening on :%s", port)
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           nil, // DefaultServeMuxを使用
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func (s *AssistantServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// リクエストボディの読み込み (io.ReadAllを使用)
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	queryText := string(bodyBytes)

	if queryText == "" {
		http.Error(w, "Empty query", http.StatusBadRequest)
		return
	}

	log.Printf("Query: %s", queryText)

	// Assistantへのリクエスト実行 (リクエストのContextを伝搬)
	responseText, err := s.sendToAssistant(r.Context(), queryText)
	if err != nil {
		log.Printf("Assistant Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte(responseText)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (s *AssistantServer) sendToAssistant(ctx context.Context, text string) (string, error) {
	stream, err := s.client.Assist(ctx)
	if err != nil {
		return "", fmt.Errorf("stream creation failed: %v", err)
	}

	req := &pb.AssistRequest{
		Type: &pb.AssistRequest_Config{
			Config: &pb.AssistConfig{
				Type: &pb.AssistConfig_TextQuery{
					TextQuery: text,
				},
				AudioOutConfig: &pb.AudioOutConfig{
					Encoding:         pb.AudioOutConfig_LINEAR16,
					SampleRateHertz:  16000,
					VolumePercentage: 0,
				},
				DeviceConfig: &pb.DeviceConfig{
					DeviceId:      s.config.DeviceID,
					DeviceModelId: s.config.ModelID,
				},
				DialogStateIn: &pb.DialogStateIn{
					LanguageCode:      s.config.Language,
					IsNewConversation: true,
				},
			},
		},
	}

	if err := stream.Send(req); err != nil {
		return "", fmt.Errorf("send request failed: %v", err)
	}

	if err := stream.CloseSend(); err != nil {
		return "", fmt.Errorf("close send failed: %v", err)
	}

	var responseBuilder string
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("receive failed: %v", err)
		}

		if res.DialogStateOut != nil {
			responseBuilder += res.DialogStateOut.SupplementalDisplayText
		}
	}

	return responseBuilder, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
