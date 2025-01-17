// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"carvel.dev/ytt/pkg/yamlmeta"
)

func (o Op) mergeMapItem(leftMap *yamlmeta.Map, newItem *yamlmeta.MapItem,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) error {

	matchChildDefaults, err := NewMatchChildDefaultsAnnotation(newItem, parentMatchChildDefaults)
	if err != nil {
		return err
	}

	ann, err := NewMapItemMatchAnnotation(newItem, parentMatchChildDefaults, o.Thread)
	if err != nil {
		return err
	}

	leftIdxs, err := ann.Indexes(leftMap)
	if err != nil {
		if err, ok := err.(MatchAnnotationNumMatchError); ok && err.isConditional() {
			return nil
		}
		return err
	}

	if len(leftIdxs) == 0 {
		// No need to traverse further
		leftMap.Items = append(leftMap.Items, newItem)
		return nil
	}

	for _, leftIdx := range leftIdxs {
		replace := true
		if leftMap.Items[leftIdx].Value != nil {
			replace, err = o.apply(leftMap.Items[leftIdx].Value, newItem.Value, matchChildDefaults)
			if err != nil {
				return err
			}
		}
		if replace {
			// left side type and metas are preserved
			err := leftMap.Items[leftIdx].SetValue(newItem.Value)
			if err != nil {
				return err
			}
			leftMap.Items[leftIdx].SetPosition(newItem.Position)
		}
	}

	return nil
}

func (o Op) removeMapItem(leftMap *yamlmeta.Map, newItem *yamlmeta.MapItem,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) error {

	ann, err := NewMapItemMatchAnnotation(newItem, parentMatchChildDefaults, o.Thread)
	if err != nil {
		return err
	}

	leftIdxs, err := ann.Indexes(leftMap)
	if err != nil {
		if err, ok := err.(MatchAnnotationNumMatchError); ok && err.isConditional() {
			return nil
		}
		return err
	}

	for _, leftIdx := range leftIdxs {
		leftMap.Items[leftIdx] = nil
	}

	// Prune out all nil items
	updatedItems := []*yamlmeta.MapItem{}

	for _, item := range leftMap.Items {
		if item != nil {
			updatedItems = append(updatedItems, item)
		}
	}

	leftMap.Items = updatedItems

	return nil
}

func (o Op) replaceMapItem(leftMap *yamlmeta.Map, newItem *yamlmeta.MapItem,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) error {

	ann, err := NewMapItemMatchAnnotation(newItem, parentMatchChildDefaults, o.Thread)
	if err != nil {
		return err
	}

	replaceAnn, err := NewReplaceAnnotation(newItem, o.Thread)
	if err != nil {
		return err
	}

	leftIdxs, err := ann.Indexes(leftMap)
	if err != nil {
		if err, ok := err.(MatchAnnotationNumMatchError); ok && err.isConditional() {
			return nil
		}
		return err
	}

	for _, leftIdx := range leftIdxs {
		newVal, err := replaceAnn.Value(leftMap.Items[leftIdx])
		if err != nil {
			return err
		}

		// left side fields are not preserved.
		// probably need to rethink how to merge left and right once those fields are needed
		leftMap.Items[leftIdx] = newItem.DeepCopy()
		err = leftMap.Items[leftIdx].SetValue(newVal)
		if err != nil {
			return err
		}
	}

	if len(leftIdxs) == 0 && replaceAnn.OrAdd() {
		newVal, err := replaceAnn.Value(nil)
		if err != nil {
			return err
		}

		leftMap.Items = append(leftMap.Items, newItem.DeepCopy())
		err = leftMap.Items[len(leftMap.Items)-1].SetValue(newVal)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o Op) assertMapItem(leftMap *yamlmeta.Map, newItem *yamlmeta.MapItem,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) error {

	matchChildDefaults, err := NewMatchChildDefaultsAnnotation(newItem, parentMatchChildDefaults)
	if err != nil {
		return err
	}

	ann, err := NewMapItemMatchAnnotation(newItem, parentMatchChildDefaults, o.Thread)
	if err != nil {
		return err
	}

	testAnn, err := NewAssertAnnotation(newItem, o.Thread)
	if err != nil {
		return err
	}

	leftIdxs, err := ann.Indexes(leftMap)
	if err != nil {
		if err, ok := err.(MatchAnnotationNumMatchError); ok && err.isConditional() {
			return nil
		}
		return err
	}

	for _, leftIdx := range leftIdxs {
		err := testAnn.Check(leftMap.Items[leftIdx])
		if err != nil {
			return err
		}

		_, err = o.apply(leftMap.Items[leftIdx].Value, newItem.Value, matchChildDefaults)
		if err != nil {
			return err
		}
	}

	return nil
}
