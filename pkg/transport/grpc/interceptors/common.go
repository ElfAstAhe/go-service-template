package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

// serverStream реализует каноничный паттерн обертки (Wrapper) над стандартным grpc.ServerStream.
//
// Применяется внутри gRPC Stream Interceptors для преодоления иммутабельности исходного потока.
// Позволяет динамически подменять и обогащать context.Context выполнения стрима (например, прокидывать
// верифицированный Subject авторизации, OTel трейсинг-спаны или уникальные RequestID) сквозь слои приложения.
type serverStream struct {
	grpc.ServerStream                 // Анонимное встраивание (Embedding) для нативного делегирования всех методов стрима
	ctx               context.Context // Мутированный, обогащенный метаданными контекст выполнения горутины
}

// Context переопределяет стандартное поведение grpc.ServerStream, возвращая кастомный
// обогащенный контекст выполнения вместо исходного сетевого контекста gRPC соединения.
func (s *serverStream) Context() context.Context {
	return s.ctx
}
