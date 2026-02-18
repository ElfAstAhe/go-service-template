package repository

// BaseQueryBuilders билдеры SQL запросов под основным методам репозитория
type BaseQueryBuilders struct {
	findBuilder   QueryBuilderFunc
	listBuilder   QueryBuilderFunc
	createBuilder QueryBuilderFunc
	changeBuilder QueryBuilderFunc
	deleteBuilder QueryBuilderFunc
}

func newBaseQueryBuilders() *BaseQueryBuilders {
	return &BaseQueryBuilders{}
}

func (bq *BaseQueryBuilders) GetFind() QueryBuilderFunc {
	return bq.findBuilder
}

func (bq *BaseQueryBuilders) GetList() QueryBuilderFunc {
	return bq.listBuilder
}

func (bq *BaseQueryBuilders) GetCreate() QueryBuilderFunc {
	return bq.createBuilder
}

func (bq *BaseQueryBuilders) GetChange() QueryBuilderFunc {
	return bq.changeBuilder
}

func (bq *BaseQueryBuilders) GetDelete() QueryBuilderFunc {
	return bq.deleteBuilder
}

// BaseQueryBuildersBuilder билдер запросов SQL
type BaseQueryBuildersBuilder struct {
	instance *BaseQueryBuilders
}

func NewBaseQueryBuildersBuilder() *BaseQueryBuildersBuilder {
	return &BaseQueryBuildersBuilder{
		instance: &BaseQueryBuilders{},
	}
}

func (bb *BaseQueryBuildersBuilder) NewInstance() *BaseQueryBuildersBuilder {
	bb.instance = newBaseQueryBuilders()

	return bb
}

func (bb *BaseQueryBuildersBuilder) WithFind(findBuilder QueryBuilderFunc) *BaseQueryBuildersBuilder {
	bb.instance.findBuilder = findBuilder

	return bb
}

func (bb *BaseQueryBuildersBuilder) WithList(listBuilder QueryBuilderFunc) *BaseQueryBuildersBuilder {
	bb.instance.listBuilder = listBuilder

	return bb
}

func (bb *BaseQueryBuildersBuilder) WithCreate(createBuilder QueryBuilderFunc) *BaseQueryBuildersBuilder {
	bb.instance.createBuilder = createBuilder

	return bb
}

func (bb *BaseQueryBuildersBuilder) WithChange(change QueryBuilderFunc) *BaseQueryBuildersBuilder {
	bb.instance.changeBuilder = change

	return bb
}

func (bb *BaseQueryBuildersBuilder) WithDelete(delete QueryBuilderFunc) *BaseQueryBuildersBuilder {
	bb.instance.deleteBuilder = delete

	return bb
}

func (bb *BaseQueryBuildersBuilder) Build() *BaseQueryBuilders {
	return bb.instance
}
