package grpc

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/knstch/shortener/cmd/config"
	"github.com/knstch/shortener/internal/app/storage/psql"
	pb "github.com/knstch/shortener/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LinksServer является интерфейсом объединяющим методы gRPC.
type LinksServer struct {
	pb.UnimplementedLinksServer

	Db *psql.PsqURLlStorage
}

var pgErr *pgconn.PgError

// FindLink ищет ссылку по короткому адресу и отдает длинную ссылку.
func (s *LinksServer) FindLink(ctx context.Context, in *pb.ShortenLink) (*pb.ShortenLinkResponse, error) {
	var response pb.ShortenLinkResponse
	longLink, deleteStatus, err := s.Db.FindLink(ctx, in.ShortenLink)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error finding link")
	}
	if deleteStatus {
		return &response, nil
	}

	response.LongLink = longLink
	response.DeleteStatus = deleteStatus

	return &response, nil
}

// PostLink записывает длинную ссылку в хранилище и отдает короткую ссылку и ошибку.
func (s *LinksServer) PostLink(ctx context.Context, in *pb.LongLink) (*pb.LongLinkResponse, error) {
	var response pb.LongLinkResponse
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	returnedShortLink, err := s.Db.PostLink(ctx, in.LongLink, config.ReadyConfig.BaseURL, userID)
	response.ShortenLink = returnedShortLink
	if err != nil {
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return &response, nil
		}
		return &response, status.Errorf(codes.Internal, "error posting link")
	}

	return &response, nil
}

// GetURLsByID получает ID клиента из куки и возвращает все ссылки отправленные им.
func (s *LinksServer) GetURLsByID(ctx context.Context, in *pb.Empty) (*pb.ListShortenLinks, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	var response pb.ListShortenLinks

	userURLs, err := s.Db.GetURLsByID(ctx, userID, config.ReadyConfig.BaseURL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error getting links")
	}

	for _, v := range userURLs {
		links := &pb.UserLinks{
			ShortLink: v.ShortLink,
			LongLink:  v.LongLink,
		}
		response.ShortenLink = append(response.ShortenLink, links)
	}
	return &response, nil
}

// DeleteURLs удаляет ссылки, отправленные клиентом при том условии, что он их загрузил.
func (s *LinksServer) DeleteURLs(ctx context.Context, in *pb.ListShortenLinksToDelete) (*pb.Empty, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	var result pb.Empty

	err = s.Db.DeleteURLs(ctx, userID, in.ShortenLink)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error deleting URLs")
	}

	return &result, nil
}

func getUserID(ctx context.Context) (int, error) {
	var id int
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get("userID")
		id, err := strconv.Atoi(values[0])
		if err != nil {
			return id, status.Errorf(codes.Internal, "can't convert to int")
		}
		if len(values) != 0 {
			return id, nil
		}
		return id, status.Errorf(codes.Unauthenticated, "invalid access token")
	}
	return id, status.Errorf(codes.DataLoss, "no metadata")
}
