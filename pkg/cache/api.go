package cache

import "fmt"

func (s *Cache) GetMembers() (members []int, err error) {
	value, err := s.gc.Get("members")
	if err != nil {
		return nil, err
	}
	if value == nil {
		return []int{}, nil
	}
	var ok bool
	members, ok = value.([]int)
	if !ok {
		return nil, fmt.Errorf("Value of key 'members' not array int, %T ", members)
	}
	return
}
