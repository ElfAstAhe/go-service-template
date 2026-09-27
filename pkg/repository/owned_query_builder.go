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

func (bq *BaseOwnedQueryBuilders) GetFind() QueryBuilderFunc {
	return bq.findBuilder
}

func (bq *BaseOwnedQueryBuilders) GetList() QueryBuilderFunc {
	return bq.listBuilder
}

func (bq *BaseOwnedQueryBuilders) GetListAll() QueryBuilderFunc {
	return bq.listAllBuilder
}

func (bq *BaseOwnedQueryBuilders) GetListAllByOwners() QueryBuilderFunc {
	return bq.listAllByOwnersBuilder
}

func (bq *BaseOwnedQueryBuilders) GetCreate() QueryBuilderFunc {
	return bq.createBuilder
}

func (bq *BaseOwnedQueryBuilders) GetChange() QueryBuilderFunc {
	return bq.changeBuilder
}

func (bq *BaseOwnedQueryBuilders) GetDeleteAll() QueryBuilderFunc {
	return bq.deleteAllBuilder
}

func (bq *BaseOwnedQueryBuilders) GetDelete() QueryBuilderFunc {
	return bq.deleteBuilder
}

// BaseOwnedQueryBuildersBuilder реализует паттерн Строитель (Builder) для пошаговой,
// безопасной сборки и конфигурации SQL-генераторов для зависимых (owned) сущностей.
type BaseOwnedQueryBuildersBuilder struct {
	instance *BaseOwnedQueryBuilders
}

// NewBaseOwnedQueryBuildersBuilder — фабричный конструктор строителя.
// Сразу аллоцирует instance в памяти, полностью защищая рантайм от паник при вызове With-методов.
func NewBaseOwnedQueryBuildersBuilder() *BaseOwnedQueryBuildersBuilder {
	return &BaseOwnedQueryBuildersBuilder{
		instance: &BaseOwnedQueryBuilders{},
	}
}

// NewInstance принудительно сбрасывает состояние билдера, подкладывая чистую структуру.
func (bbo *BaseOwnedQueryBuildersBuilder) NewInstance() *BaseOwnedQueryBuildersBuilder {
	bbo.instance = newBaseOwnerQueryBuilders()

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithFind(findBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.findBuilder = findBuilder

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithList(listBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.listBuilder = listBuilder

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithListAll(listAllBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.listAllBuilder = listAllBuilder

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithListAllByOwners(listAllByOwnersBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.listAllByOwnersBuilder = listAllByOwnersBuilder

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithCreate(createBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.createBuilder = createBuilder

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithChange(change QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.changeBuilder = change

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithDelete(deleteBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.deleteBuilder = deleteBuilder

	return bbo
}

func (bbo *BaseOwnedQueryBuildersBuilder) WithDeleteAll(deleteAllBuilder QueryBuilderFunc) *BaseOwnedQueryBuildersBuilder {
	bbo.instance.deleteAllBuilder = deleteAllBuilder

	return bbo
}

// Build завершает конфигурацию строителя и возвращает готовый пул SQL-генераторов.
func (bbo *BaseOwnedQueryBuildersBuilder) Build() *BaseOwnedQueryBuilders {
	return bbo.instance
}
