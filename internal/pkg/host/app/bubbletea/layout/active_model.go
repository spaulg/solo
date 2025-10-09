package layout

type ActiveModel struct {
	activeModel *string
}

func NewActiveModel() *ActiveModel {
	return &ActiveModel{}
}

func (t *ActiveModel) SetActiveModel(model string) {
	t.activeModel = &model
}

func (t *ActiveModel) IsActive(model string) bool {
	return *(t.activeModel) == model
}
