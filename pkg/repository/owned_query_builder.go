package repository

// BaseOwnedQueryBuilders инкапсулирует в себе набор декларативных функций (колбеков),
// генерирующих строки SQL-запросов для основных операций Owned-репозитория.
//
// Обеспечивает полную изоляцию текстовых SQL-шаблонов (включая сложные Bulk Select и IN-выборки)
// от логики их выполнения инфраструктурным движком OwnedHelper.
type BaseOwnedQueryBuilders struct {
	findBuilder            QueryBuilderFunc // SELECT BY ID с проверкой owner_id
	listBuilder            QueryBuilderFunc // SELECT LIST с пагинацией и проверкой owner_id
	listAllBuilder         QueryBuilderFunc // SELECT ALL для полной выгрузки коллекции владельца
	listAllByOwnersBuilder QueryBuilderFunc // BULK SELECT для пакетной IN-выборки по группе владельцев
	createBuilder          QueryBuilderFunc // INSERT с привязкой к owner_id
	changeBuilder          QueryBuilderFunc // UPDATE с проверкой owner_id
	deleteAllBuilder       QueryBuilderFunc // CASCADE DELETE по owner_id
	deleteBuilder          QueryBuilderFunc // DELETE одиночной записи по ID и owner_id
}

// newBaseOwnerQueryBuilders инициализирует пустую структуру билдеров для owned-сущностей.
func newBaseOwnerQueryBuilders() *BaseOwnedQueryBuilders {
	return &BaseOwnedQueryBuilders{}
}

// ====================================================================
// Публичные геттеры для безопасного извлечения функций генерации SQL
// ====================================================================

// GetFind возвращает функцию-генератор SQL-строки для операции точечной выборки зависимой сущности по ее ID и идентификатору владельца (SELECT BY ID).
func (bq *BaseOwnedQueryBuilders) GetFind() QueryBuilderFunc {
	return bq.findBuilder
}

// GetList возвращает функцию-генератор SQL-строки для постраничной выборки зависимых сущностей конкретного владельца (SELECT LIST).
func (bq *BaseOwnedQueryBuilders) GetList() QueryBuilderFunc {
	return bq.listBuilder
}

// GetListAll возвращает функцию-генератор SQL-строки для полной выгрузки всей коллекции дочерних записей указанного владельца (SELECT ALL).
func (bq *BaseOwnedQueryBuilders) GetListAll() QueryBuilderFunc {
	return bq.listAllBuilder
}

// GetListAllByOwners возвращает функцию-генератор SQL-строки для пакетного bulk-извлечения данных по группе владельцев за один запрос (IN-выборка).
func (bq *BaseOwnedQueryBuilders) GetListAllByOwners() QueryBuilderFunc {
	return bq.listAllByOwnersBuilder
}

// GetCreate возвращает функцию-генератор SQL-строки для операции вставки новой дочерней записи с жесткой привязкой к владельцу (INSERT).
func (bq *BaseOwnedQueryBuilders) GetCreate() QueryBuilderFunc {
	return bq.createBuilder
}

// GetChange возвращает функцию-генератор SQL-строки для операции модификации полей существующей owned-записи (UPDATE).
func (bq *BaseOwnedQueryBuilders) GetChange() QueryBuilderFunc {
	return bq.changeBuilder
}

// GetDeleteAll возвращает функцию-генератор SQL-строки для каскадного удаления всей коллекции дочерних ресурсов указанного владельца (CASCADE DELETE).
func (bq *BaseOwnedQueryBuilders) GetDeleteAll() QueryBuilderFunc {
	return bq.deleteAllBuilder
}

// GetDelete возвращает функцию-генератор SQL-строки для точечного удаления связанного дочернего ресурса по его ID и owner_id (DELETE).
func (bq *BaseOwnedQueryBuilders) GetDelete() QueryBuilderFunc {
	return bq.deleteBuilder
}

// BaseOwnedQueryBuildersBuilder реализует паттерн Строитель (Builder) для пошаговой,
// безопасной сборки и конфигурации SQL-генераторов для зависимых (owned) сущностей (Fluent Builder API).
type BaseOwnedQueryBuildersBuilder struct {
	instance *BaseOwnedQueryBuilders
}

// NewBaseOwnedQueryBuildersBuilder — фабричный конструктор строителя.
// Сразу аллоцирует instance в памяти, полностью защищая рантайм от паник при вызове With-методов (Guard Clause).
func NewBaseOwnedQueryBuildersBuilder() *BaseOwnedQueryBuildersBuilder {
	return &BaseOwnedQueryBuildersBuilder{
		instance: &BaseOwnedQueryBuilders{},
	}
}

// NewInstance принудительно сбрасывает состояние билдера, подкладывая чистую структуру (мутатор повторного использования).
func (bbo *BaseOwnedQueryBuildersBuilder) NewInstance() *BaseOwnedQueryBuildersBuilder {
	bbo.instance = newBaseOwnerQueryBuilders()

	return bbo
}

// WithFind инжектирует в собираемый объект функцию генерации SQL-запроса SELECT BY ID с проверкой owner_id.
func (bbo *BaseOwnedQueryBuildersBuilder) WithFind(findBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.findBuilder = findBuilder

	return bbo
}

// WithList инжектирует в собираемый объект функцию генерации SQL-запроса SELECT LIST с пагинацией и проверкой owner_id.
func (bbo *BaseOwnedQueryBuildersBuilder) WithList(listBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.listBuilder = listBuilder

	return bbo
}

// WithListAll инжектирует в собираемый объект функцию генерации SQL-запроса SELECT ALL для полной выгрузки коллекции владельца.
func (bbo *BaseOwnedQueryBuildersBuilder) WithListAll(listAllBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.listAllBuilder = listAllBuilder

	return bbo
}

// WithListAllByOwners инжектирует в собираемый объект функцию генерации SQL-запроса bulk-выборки по группе владельцев (IN-запрос).
func (bbo *BaseOwnedQueryBuildersBuilder) WithListAllByOwners(listAllByOwnersBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.listAllByOwnersBuilder = listAllByOwnersBuilder

	return bbo
}

// WithCreate инжектирует в собираемый объект функцию генерации SQL-запроса INSERT с привязкой к owner_id.
func (bbo *BaseOwnedQueryBuildersBuilder) WithCreate(createBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.createBuilder = createBuilder

	return bbo
}

// WithChange инжектирует в собираемый объект функцию генерации SQL-запроса UPDATE с проверкой owner_id.
func (bbo *BaseOwnedQueryBuildersBuilder) WithChange(change QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.changeBuilder = change

	return bbo
}

// WithDelete инжектирует в собираемый объект функцию генерации SQL-запроса DELETE одиночной записи по ID и owner_id.
func (bbo *BaseOwnedQueryBuildersBuilder) WithDelete(deleteBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.deleteBuilder = deleteBuilder

	return bbo
}

// WithDeleteAll инжектирует в собираемый объект функцию генерации SQL-запроса CASCADE DELETE по owner_id.
func (bbo *BaseOwnedQueryBuildersBuilder) WithDeleteAll(deleteAllBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.deleteAllBuilder = deleteAllBuilder

	return bbo
}

// Build завершает пошаговую конфигурацию строителя и возвращает полностью укомплектованный пул SQL-генераторов зависимых сущностей.
func (bbo *BaseOwnedQueryBuildersBuilder) Build() *BaseOwnedQueryBuilders {
	return bbo.instance
}
