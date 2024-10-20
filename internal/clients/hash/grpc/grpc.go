package grpc

import (
	"context"
	"fmt"
	hashv1 "github.com/Chugan908/hash-contracts/gen/go"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"time"
)

type Client struct {
	api hashv1.HashClient
	log *slog.Logger
}

func New(
	ctx context.Context,
	log *slog.Logger,
	addr string,
	timeout time.Duration,
	retriesCount int,
) (*Client, error) {
	const op = "grpc.New"

	// If client receives one of described error codes, it will repeat the request

	retyr0pts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),  // amount of retries
		grpcretry.WithPerRetryTimeout(timeout), // timeout between them
	}

	log0pts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent), // we log the body of request and the body of response
	}

	// An Interceptor is a function that is invoked by the framework BEFORE or AFTER an action invocation.
	cc, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), log0pts...),
			grpcretry.UnaryClientInterceptor(retyr0pts...),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &Client{
		api: hashv1.NewHashClient(cc),
		log: log,
	}, nil
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields)
	})
}

func (c *Client) HashUrl(ctx context.Context, textUrl string) (string, error) {
	const op = "grpc.HashUrl"

	resp, err := c.api.HashUrl(ctx, &hashv1.HashUrlRequest{
		TextUrl: textUrl,
	})
	if err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}

	return resp.GetTexthashUrl(), nil
}
