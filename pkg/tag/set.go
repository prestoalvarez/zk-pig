package tag

var EmptySet = Set{}

// Set reprensents a immutable set of
type Set []*Tag

// WithTags returns a new set with the given tags added to the set. If a tag with the same key already exists in the set,
// the new tag will replace the old tag.
func (s Set) WithTags(tags ...*Tag) Set {
	copySet := make(Set, len(s))
	for i, t := range s {
		copySet[i] = t.Copy()
	}

	newTags := make(Set, 0)
	for _, tag := range tags {
		var existed bool
		for i, oldTag := range s {
			if oldTag.Key == tag.Key {
				copySet[i] = merge(copySet[i], tag)
				existed = true
				break
			}
		}
		if !existed {
			newTags = append(newTags, tag)
		}
	}

	return append(copySet, newTags...)
}

func merge(cpyTag, newTag *Tag) *Tag {
	if cpyTag.chained != nil && *cpyTag.chained {
		if cpyTag.Value.Type == STRING {
			cpyTag.Value.Interface = cpyTag.Value.Interface.(string) + "." + newTag.Value.Interface.(string)
			// If the newtag has a chain flag set, we propagate it to the cpyTag
			if newTag.chained != nil {
				cpyTag.chained = newTag.chained
			}
			return cpyTag
		}
		panic("cannot chain non-string values")
	}
	return newTag
}
