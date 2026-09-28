// Package telemetry реализует конкретные инфраструктурные адаптеры и провайдеры
// подсистемы распределенной наблюдаемости OpenTelemetry (OpenTelemetry Infrastructure Providers).
//
// Отвечает за низкоуровневое конфигурирование OTLP-экспортеров (gRPC/HTTP), сборку и
// регистрацию глобальных провайдеров трассировок (TracerProvider) и каскадный сброс (Flush)
// буферов сокетов со спанами при Graceful Shutdown приложения, материализуя контракты ядра.
package telemetry
