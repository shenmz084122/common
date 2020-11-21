package grpcwrap

import (
	"context"

	"github.com/DataWorkbench/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// loggerUnaryServerInterceptor create an new logger with requestId.
// You can get logger by glog.FromContext(cxt) after.
func loggerUnaryServerInterceptor(lp *glog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// Copy a new logger
		nl := lp.Clone()
		nl.WithFields().AddString(ctxReqIdKey, reqIdFromIncomingContext(ctx))

		ctx = glog.WithContext(ctx, nl)
		resp, err = handler(ctx, req)

		// Close the logger instances
		_ = nl.Close()
		return resp, err
	}
}

// recoverUnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func recoverUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		panicked := true
		defer func() {
			if r := recover(); r != nil || panicked {
				// TODO: dump runtime stack when panic
				glog.FromContext(ctx).Error().Any("unary server panic recover", r).Fire()
				err = status.Errorf(codes.Internal, "unary server panic recover: %v", r)
			}
		}()

		resp, err = handler(ctx, req)
		panicked = false
		return
	}
}

// basicUnaryServerInterceptor do validate the argument and print log
func basicUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger := glog.FromContext(ctx)

		logger.Debug().String("receive method", info.FullMethod).RawString("request", pbMsgToString(logger, req)).Fire()

		// Validated request parameters
		if err := validateRequestArgument(req, logger); err != nil {
			return nil, err
		}

		reply, err := handler(ctx, req)

		if err != nil {
			logger.Error().Error("handled with error", err).Fire()
			return nil, err
		}

		logger.Debug().RawString("handled with reply", pbMsgToString(logger, reply)).Fire()
		return reply, err
	}
}
