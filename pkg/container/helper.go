package container

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// Глобальное состояние рантайма IoC/DI контейнеризации приложения.
var (
	// defaultOrchestrator — центральный синглтон-реестр, управляющий графом всех зависимостей.
	defaultOrchestrator Orchestrator
)

// SetDefaultOrchestrator выполняет потоковую инжекцию глобального оркестратора при старте приложения.
func SetDefaultOrchestrator(orc Orchestrator) {
	defaultOrchestrator = orc
}

// GetInstance — универсальный строго типизированный (Generic) локатор служб (Service Locator).
//
// Выполняет сквозной обход всех зарегистрированных контейнеров оркестратора, находит компонент
// по его строковому имени (name), безопасно приводит к типу T и возвращает готовый объект.
// Возвращает типизированные Container errors, полностью защищая рантайм от паник несоответствия типов.
func GetInstance[T any](name string) (T, error) {
	var nilRes T
	// Предохранитель: проверяем, что глобальный оркестратор инициализирован в main.go
	if utils.IsNil(defaultOrchestrator) {
		return nilRes, errs.NewContainerValidateError("default orchestrator", "GetInstance", "default orchestrator not set up", nil)
	}

	var err error
	var res T
	// Каскадно сканируем граф контейнеров для поиска целевого компонента
	for _, cnt := range defaultOrchestrator.AllContainers() {
		if cnt.IsRegistered(name) {
			// Требует уникального именования инстансов в рамках всего приложения во избежание коллизий
			res, err = GetContainerInstance[T](cnt, name)
			if err != nil {
				return nilRes, err
			}
			if !utils.IsNil(res) {
				return res, nil
			}
		}
	}

	return nilRes, errs.NewContainerNotFoundError(fmt.Sprintf("instance %s not found in all registered containers", name), nil)
}

// GetContainerInstance выполняет точечное извлечение объекта из конкретного изолированного контейнера
// с проведением валидации контрактов и безопасного динамического приведения типов (Type Assertion).
func GetContainerInstance[T any](container Container, name string) (T, error) {
	var nilRes T

	// Шаг 1: Валидация входных аргументов на предмет пустых строк или nil-указателей
	if err := getInstanceValidate(container, name); err != nil {
		return nilRes, err
	}

	// Шаг 2: Извлекаем нетипизированный интерфейс any из внутреннего реестра контейнера
	instance, err := container.GetInstance(name)
	if err != nil {
		return nilRes, err
	}
	if utils.IsNil(instance) {
		return nilRes, nil
	}

	// Шаг 3: Безопасный динамический кастинг интерфейса к целевому дженерик-типу T
	res, ok := instance.(T)
	if !ok {
		// В случае несовпадения типов генерируем детальную ошибку с выводом полного пути структуры (Full Type Name)
		return nilRes, errs.NewContainerError(container.GetName(), fmt.Sprintf("instance type [%s] mismatch", utils.GetFullTypeName(instance)), nil)
	}

	return res, nil
}

// getInstanceValidate — внутренний оборонительный предохранитель (Guard), пресекающий некорректные вызовы DI.
func getInstanceValidate(container Container, name string) error {
	if container == nil {
		return errs.NewContainerValidateError("container", "GetContainerInstance", "container nil", nil)
	}
	if name == "" {
		return errs.NewContainerValidateError(container.GetName(), "GetContainerInstance", "instance name is empty", nil)
	}

	return nil
}
