package goreflect

import (
    "errors"
    "reflect"
)

// SliceGroup 对已经是group by结果出来的sql结果，进行分组处理
type SliceGroup struct {
    start, end int                           //  当前分组的开始结束索引
    group      interface{}                   // 当前分组值
    slice      reflect.Value                 // 原始数组
    len        int                           // 原始数组长度
    byFn       func(interface{}) interface{} // 从数组元素中获取分组值的函数
}

// MakeSliceGroup 创建SliceGroup对象
func MakeSliceGroup(slice interface{}, byFn func(interface{}) interface{}) (*SliceGroup, error) {
    v := reflect.ValueOf(slice)
    switch v.Kind() {
    case reflect.Slice, reflect.Array: // ok
    default:
        return nil, errors.New("first argument should slice/array")
    }

    return &SliceGroup{slice: v, byFn: byFn, len: v.Len()}, nil
}

// NextGroup 返回下一个分组的group和slice
func (s *SliceGroup) NextGroup() (group interface{}, groupSlice interface{}, ok bool) {
    started := false
    for s.start = s.end; s.end < s.len; s.end++ {
        gr := s.byFn(s.slice.Index(s.end).Interface())
        if !started {
            started = true
            s.group = gr
            continue
        }

        if s.group != gr {
            break
        }
    }

    if s.start < s.end {
        return s.group, s.slice.Slice(s.start, s.end).Interface(), true
    }

    return "", nil, false
}
