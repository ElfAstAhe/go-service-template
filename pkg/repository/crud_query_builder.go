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

func (bq *BaseCRUDQueryBuilders) GetFind() QueryBuilderFunc {
	return bq.findBuilder
}

func (bq *BaseCRUDQueryBuilders) GetList() QueryBuilderFunc {
	return bq.listBuilder
}

func (bq *BaseCRUDQueryBuilders) GetCreate() QueryBuilderFunc {
	return bq.createBuilder
}

func (bq *BaseCRUDQueryBuilders) GetChange() QueryBuilderFunc {
	return bq.changeBuilder
}

func (bq *BaseCRUDQueryBuilders) GetDelete() QueryBuilderFunc {
	return bq.deleteBuilder
}

// BaseCRUDQueryBuildersBuilder реализует паттерн Строитель (Builder) для пошаговой,
// безопасной сборки и конфигурации SQL-генераторов сущности.
type BaseCRUDQueryBuildersBuilder struct {
	instance *BaseCRUDQueryBuilders
}

// NewBaseCRUDQueryBuildersBuilder — фабричный конструктор строителя.
// Сразу аллоцирует instance в памяти, полностью защищая рантайм от паник при вызове With-методов.
func NewBaseCRUDQueryBuildersBuilder() *BaseCRUDQueryBuildersBuilder {
	return &BaseCRUDQueryBuildersBuilder{
		instance: &BaseCRUDQueryBuilders{},
	}
}

// NewInstance принудительно сбрасывает состояние билдера, подкладывая чистую структуру.
func (bb *BaseCRUDQueryBuildersBuilder) NewInstance() *BaseCRUDQueryBuildersBuilder {
	bb.instance = newBaseCRUDQueryBuilders()

	return bb
}

func (bb *BaseCRUDQueryBuildersBuilder) WithFind(findBuilder QueryBuilderFunc) *BaseCRUDQueryBuildersBuilder {
	bb.instance.findBuilder = findBuilder

	return bb
}

func (bb *BaseCRUDQueryBuildersBuilder) WithList(listBuilder QueryBuilderFunc) *BaseCRUDQueryBuildersBuilder {
	bb.instance.listBuilder = listBuilder

	return bb
}

func (bb *BaseCRUDQueryBuildersBuilder) WithCreate(createBuilder QueryBuilderFunc) *BaseCRUDQueryBuildersBuilder {
	bb.instance.createBuilder = createBuilder

	return bb
}

func (bb *BaseCRUDQueryBuildersBuilder) WithChange(change QueryBuilderFunc) *BaseCRUDQueryBuildersBuilder {
	bb.instance.changeBuilder = change

	return bb
}

func (bb *BaseCRUDQueryBuildersBuilder) WithDelete(delete QueryBuilderFunc) *BaseCRUDQueryBuildersBuilder {
	bb.instance.deleteBuilder = delete

	return bb
}

// Build завершает конфигурацию строителя и возвращает готовый пул SQL-генераторов.
func (bb *BaseCRUDQueryBuildersBuilder) Build() *BaseCRUDQueryBuilders {
	return bb.instance
}
