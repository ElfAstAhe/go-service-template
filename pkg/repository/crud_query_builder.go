package repository

// BaseCRUDQueryBuilders билдеры SQL запросов под основным методам CRUD репозитория
type BaseCRUDQueryBuilders struct {
	findBuilder   QueryBuilderFunc
	listBuilder   QueryBuilderFunc
	createBuilder QueryBuilderFunc
	changeBuilder QueryBuilderFunc
	deleteBuilder QueryBuilderFunc
}

func newBaseCRUDQueryBuilders() *BaseCRUDQueryBuilders {
	return &BaseCRUDQueryBuilders{}
}

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

// BaseCRUDQueryBuildersBuilder билдер запросов SQL
type BaseCRUDQueryBuildersBuilder struct {
	instance *BaseCRUDQueryBuilders
}

func NewBaseCRUDQueryBuildersBuilder() *BaseCRUDQueryBuildersBuilder {
	return &BaseCRUDQueryBuildersBuilder{
		instance: &BaseCRUDQueryBuilders{},
	}
}

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

func (bb *BaseCRUDQueryBuildersBuilder) Build() *BaseCRUDQueryBuilders {
	return bb.instance
}
