package data

type TagCache struct {
	tags []*Tag
}

func NewTagCache() *TagCache {
	return &TagCache{}
}

func (t *TagCache) GetByName(name string) *Tag {
	for _, tag := range t.tags {
		if tag.Name == name {
			return tag
		}
	}
	return nil
}

func (t *TagCache) add(tag *Tag) {
	t.tags = append(t.tags, tag)
}

func (t *TagCache) getById(id int) *Tag {
	for _, tag := range t.tags {
		if tag.Id == id {
			return tag
		}
	}
	return nil
}
