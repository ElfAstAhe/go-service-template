package repository

// BaseCRUDQueryBuilders инкапсулирует в себе набор декларативных функций (колбеков),
// генерирующих строки SQL-запросов для основных операций CRUD-репозитория.
//
// Позволяет полностью изолировать текстовые SQL-шаблоны от логики их выполнения.
type BaseCRUDQueryBuilders struct {
	findBuilder   QueryBuilderFunc // Генератор SQL для выборки по ID (SELECT BY ID)
	listBuilder   QueryBuilderFunc // Генератор SQL для пагинации (SELECT LIST)
	createBuilder QueryBuilderFunc // Генератор SQL для вставки (INSERT)
	changeBuilder QueryBuilderFunc // Генератор SQL для модификации (UPDATE)
	deleteBuilder QueryBuilderFunc // Генератор SQL для удаления (DELETE)
}

// newBaseCRUDQueryBuilders инициализирует пустую структуру билдеров.
func newBaseCRUDQueryBuilders() *BaseCRUDQueryBuilders {
	return &BaseCRUDQueryBuilders{}
}

// ====================================================================
// Публичные геттеры для безопасного извлечения функций генерации SQL
// ====================================================================

// GetFind возвращает функцию-генератор SQL-строки для операции точечной выборки по идентификатору (SELECT BY ID).
func (bq *BaseCRUDQueryBuilders) GetFind() QueryBuilderFunc {
	return bq.findBuilder
}

// GetList returns the SQL string generator function for paginated list selection operations (SELECT LIST).
func (bq *BaseCRUDQueryBuilders) GetList() QueryBuilderFunc {
	return bq.listBuilder
}

// GetCreate возвращает функцию-генератор SQL-строки для операции вставки новой записи (INSERT).
func (bq *BaseCRUDQueryBuilders) GetCreate() QueryBuilderFunc {
	return bq.createBuilder
}

// GetChange возвращает функцию-генератор SQL-строки для операции модификации полей существующей записи (UPDATE).
func (bq *BaseCRUDQueryBuilders) GetChange() QueryBuilderFunc {
	return bq.changeBuilder
}

// GetDelete возвращает функцию-генератор SQL-строки для операции удаления записи из СУБД (DELETE).
func (bq *BaseCRUDQueryBuilders) GetDelete() QueryBuilderFunc {
	return bq.deleteBuilder
}

// BaseCRUDQueryBuildersBuilder реализует паттерн Строитель (Builder) для пошаговой,
// безопасной сборки и конфигурации SQL-генераторов сущности (Fluent Builder API).
type BaseCRUDQueryBuildersBuilder struct {
	instance *BaseCRUDQueryBuilders
}

// NewBaseCRUDQueryBuildersBuilder — фабричный конструктор строителя.
// Сразу аллоцирует instance в памяти, полностью защищая рантайм от паник при вызове With-методов (Guard Clause).
func NewBaseCRUDQueryBuildersBuilder() *BaseCRUDQueryBuildersBuilder {
	return &BaseCRUDQueryBuildersBuilder{
		instance: &BaseCRUDQueryBuilders{},
	}
}

// NewInstance принудительно сбрасывает состояние билдера, подкладывая чистую структуру (мутатор повторного использования).
func (bb *BaseCRUDQueryBuildersBuilder) NewInstance() *BaseCRUDQueryBuildersBuilder {
	bb.instance = newBaseCRUDQueryBuilders()

	return bb
}

// WithFind инжектирует в собираемый объект функцию генерации SQL-запроса SELECT BY ID.
func (bb *BaseCRUDQueryBuildersBuilder) WithFind(findBuilder QueryBuilderFunc) *BaseCRUDQueryBuildersBuilder {
	bb.instance.findBuilder = findBuilder

	return bb
}

// WithList инжектирует в собираемый объект функцию генерации SQL-запроса SELECT LIST с поддержкой пагинации.
func (bb *BaseCRUDQueryBuildersBuilder) WithList(listBuilder QueryBuilderFunc) *BaseCRUDQueryBuildersBuilder {
	bb.instance.listBuilder = listBuilder

	return bb
}

// WithCreate инжектирует в собираемый объект функцию генерации SQL-запроса INSERT.
func (bb *BaseCRUDQueryBuildersBuilder) WithCreate(createBuilder QueryBuilderFunc) *BaseCRUDQueryBuildersBuilder {
	bb.instance.createBuilder = createBuilder

	return bb
}

// WithChange инжектирует в собираемый объект функцию генерации SQL-запроса UPDATE.
func (bb *BaseCRUDQueryBuildersBuilder) WithChange(change QueryBuilderFunc) *BaseCRUDQueryBuildersBuilder {
	bb.instance.changeBuilder = change

	return bb
}

// WithDelete инжектирует в собираемый объект функцию генерации SQL-запроса DELETE.
func (bb *BaseCRUDQueryBuildersBuilder) WithDelete(delete QueryBuilderFunc) *BaseCRUDQueryBuildersBuilder {
	bb.instance.deleteBuilder = delete

	return bb
}

// Build завершает пошаговую конфигурацию строителя и возвращает полностью укомплектованный пул SQL-генераторов.
func (bb *BaseCRUDQueryBuildersBuilder) Build() *BaseCRUDQueryBuilders {
	return bb.instance
}
