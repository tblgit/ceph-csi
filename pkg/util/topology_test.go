/*
Copyright 2020 ceph-csi authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"fmt"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

// nolint: gocyclo
// TestFindPoolAndTopology also tests MatchTopologyForPool
func TestFindPoolAndTopology(t *testing.T) {
	var err error
	var label1 = "region"
	var label2 = "zone"
	var l1Value1 = "R1"
	var l1Value2 = "R2"
	var l2Value1 = "Z1"
	var l2Value2 = "Z2"
	var pool1 = "PoolA"
	var pool2 = "PoolB"
	var topologyPrefix = "prefix"
	var emptyTopoPools = []TopologyConstrainedPool{}
	var emptyPoolNameTopoPools = []TopologyConstrainedPool{
		{
			DomainSegments: []topologySegment{
				{
					DomainLabel: label1,
					DomainValue: l1Value1,
				},
				{
					DomainLabel: label2,
					DomainValue: l2Value1,
				},
			},
		},
	}
	var emptyDomainsInTopoPools = []TopologyConstrainedPool{
		{
			PoolName: pool1,
		},
	}
	var partialDomainsInTopoPools = []TopologyConstrainedPool{
		{
			PoolName: pool1,
			DomainSegments: []topologySegment{
				{
					DomainLabel: label1,
					DomainValue: l1Value1,
				},
			},
		},
	}
	var differentDomainsInTopoPools = []TopologyConstrainedPool{
		{
			PoolName: pool1,
			DomainSegments: []topologySegment{
				{
					DomainLabel: label1 + "fuzz1",
					DomainValue: l1Value1,
				},
				{
					DomainLabel: label2,
					DomainValue: l2Value1,
				},
			},
		},
		{
			PoolName: pool2,
			DomainSegments: []topologySegment{
				{
					DomainLabel: label1,
					DomainValue: l1Value2,
				},
				{
					DomainLabel: label2,
					DomainValue: l2Value2 + "fuzz1",
				},
			},
		},
	}
	var validSingletonTopoPools = []TopologyConstrainedPool{
		{
			PoolName: pool1,
			DomainSegments: []topologySegment{
				{
					DomainLabel: label1,
					DomainValue: l1Value1,
				},
				{
					DomainLabel: label2,
					DomainValue: l2Value1,
				},
			},
		},
	}
	var validMultipleTopoPools = []TopologyConstrainedPool{
		{
			PoolName: pool1,
			DomainSegments: []topologySegment{
				{
					DomainLabel: label1,
					DomainValue: l1Value1,
				},
				{
					DomainLabel: label2,
					DomainValue: l2Value1,
				},
			},
		},
		{
			PoolName: pool2,
			DomainSegments: []topologySegment{
				{
					DomainLabel: label1,
					DomainValue: l1Value2,
				},
				{
					DomainLabel: label2,
					DomainValue: l2Value2,
				},
			},
		},
	}
	var emptyAccReq = csi.TopologyRequirement{}
	var emptySegmentAccReq = csi.TopologyRequirement{
		Requisite: []*csi.Topology{
			{},
			{},
		},
	}
	var partialHigherSegmentAccReq = csi.TopologyRequirement{
		Preferred: []*csi.Topology{
			{
				Segments: map[string]string{
					topologyPrefix + "/" + label1: l1Value1,
				},
			},
		},
	}
	var partialLowerSegmentAccReq = csi.TopologyRequirement{
		Preferred: []*csi.Topology{
			{
				Segments: map[string]string{
					topologyPrefix + "/" + label2: l2Value1,
				},
			},
		},
	}
	var differentSegmentAccReq = csi.TopologyRequirement{
		Requisite: []*csi.Topology{
			{
				Segments: map[string]string{
					topologyPrefix + "/" + label1 + "fuzz2": l1Value1,
					topologyPrefix + "/" + label2:           l2Value1,
				},
			},
			{
				Segments: map[string]string{
					topologyPrefix + "/" + label1: l1Value2,
					topologyPrefix + "/" + label2: l2Value2 + "fuzz2",
				},
			},
		},
	}
	var validAccReq = csi.TopologyRequirement{
		Requisite: []*csi.Topology{
			{
				Segments: map[string]string{
					topologyPrefix + "/" + label1: l1Value1,
					topologyPrefix + "/" + label2: l2Value1,
				},
			},
			{
				Segments: map[string]string{
					topologyPrefix + "/" + label1: l1Value2,
					topologyPrefix + "/" + label2: l2Value2,
				},
			},
		},
		Preferred: []*csi.Topology{
			{
				Segments: map[string]string{
					topologyPrefix + "/" + label1: l1Value1,
					topologyPrefix + "/" + label2: l2Value1,
				},
			},
			{
				Segments: map[string]string{
					topologyPrefix + "/" + label1: l1Value2,
					topologyPrefix + "/" + label2: l2Value2,
				},
			},
		},
	}

	checkOutput := func(err error, poolName string, topoSegment map[string]string) error {
		if err != nil {
			return fmt.Errorf("expected success, got err (%v)", err)
		}
		if poolName != pool1 || !(len(topoSegment) == 2) &&
			topoSegment[topologyPrefix+"/"+label1] == l1Value1 &&
			topoSegment[topologyPrefix+"/"+label2] == l2Value1 {
			return fmt.Errorf("expected poolName (%s) and topoSegment (%s %s), got (%s) and (%v)", pool1,
				topologyPrefix+"/"+label1+l1Value1, topologyPrefix+"/"+label2+l2Value1,
				poolName, topoSegment)
		}
		return nil
	}
	// Test nil values
	_, _, err = FindPoolAndTopology(nil, nil)
	if err != nil {
		t.Errorf("expected success due to nil in-args (%v)", err)
	}

	poolName, _, err := FindPoolAndTopology(&validMultipleTopoPools, nil)
	if err != nil || poolName != "" {
		t.Errorf("expected success due to nil accessibility requirements (err - %v) (poolName - %s)", err, poolName)
	}

	poolName, _, err = FindPoolAndTopology(nil, &validAccReq)
	if err != nil || poolName != "" {
		t.Errorf("expected success due to nil topology pools (err - %v) (poolName - %s)", err, poolName)
	}

	// Test valid accessibility requirement, with invalid topology pools values
	_, _, err = FindPoolAndTopology(&emptyTopoPools, &validAccReq)
	if err == nil {
		t.Errorf("expected failure due to empty topology pools")
	}

	_, _, err = FindPoolAndTopology(&emptyPoolNameTopoPools, &validAccReq)
	if err == nil {
		t.Errorf("expected failure due to missing pool name in topology pools")
	}

	_, _, err = FindPoolAndTopology(&differentDomainsInTopoPools, &validAccReq)
	if err == nil {
		t.Errorf("expected failure due to mismatching domains in topology pools")
	}

	// Test valid topology pools, with invalid accessibility requirements
	_, _, err = FindPoolAndTopology(&validMultipleTopoPools, &emptyAccReq)
	if err == nil {
		t.Errorf("expected failure due to empty accessibility requirements")
	}

	_, _, err = FindPoolAndTopology(&validSingletonTopoPools, &emptySegmentAccReq)
	if err == nil {
		t.Errorf("expected failure due to empty segments in accessibility requirements")
	}

	_, _, err = FindPoolAndTopology(&validMultipleTopoPools, &partialHigherSegmentAccReq)
	if err == nil {
		t.Errorf("expected failure due to partial segments in accessibility requirements")
	}

	_, _, err = FindPoolAndTopology(&validSingletonTopoPools, &partialLowerSegmentAccReq)
	if err == nil {
		t.Errorf("expected failure due to partial segments in accessibility requirements")
	}

	_, _, err = FindPoolAndTopology(&validMultipleTopoPools, &partialLowerSegmentAccReq)
	if err == nil {
		t.Errorf("expected failure due to partial segments in accessibility requirements")
	}

	_, _, err = FindPoolAndTopology(&validMultipleTopoPools, &differentSegmentAccReq)
	if err == nil {
		t.Errorf("expected failure due to mismatching segments in accessibility requirements")
	}

	// Test success cases
	// If a pool is a superset of domains (either empty domain labels or partial), it can be selected
	poolName, topoSegment, err := FindPoolAndTopology(&emptyDomainsInTopoPools, &validAccReq)
	err = checkOutput(err, poolName, topoSegment)
	if err != nil {
		t.Errorf("expected success got: (%v)", err)
	}

	poolName, topoSegment, err = FindPoolAndTopology(&partialDomainsInTopoPools, &validAccReq)
	err = checkOutput(err, poolName, topoSegment)
	if err != nil {
		t.Errorf("expected success got: (%v)", err)
	}

	// match in a singleton topology pools
	poolName, topoSegment, err = FindPoolAndTopology(&validSingletonTopoPools, &validAccReq)
	err = checkOutput(err, poolName, topoSegment)
	if err != nil {
		t.Errorf("expected success got: (%v)", err)
	}

	// match first in multiple topology pools
	poolName, topoSegment, err = FindPoolAndTopology(&validMultipleTopoPools, &validAccReq)
	err = checkOutput(err, poolName, topoSegment)
	if err != nil {
		t.Errorf("expected success got: (%v)", err)
	}

	// match non-first in multiple topology pools
	switchPoolOrder := []TopologyConstrainedPool{}
	switchPoolOrder = append(switchPoolOrder, validMultipleTopoPools[1], validMultipleTopoPools[0])
	poolName, topoSegment, err = FindPoolAndTopology(&switchPoolOrder, &validAccReq)
	err = checkOutput(err, poolName, topoSegment)
	if err != nil {
		t.Errorf("expected success got: (%v)", err)
	}

	// TEST: MatchTopologyForPool
	// check for non-existent pool
	_, err = MatchTopologyForPool(&validMultipleTopoPools, &validAccReq, pool1+"fuzz")
	if err == nil {
		t.Errorf("expected failure due to non-existent pool name (%s) got success", pool1+"fuzz")
	}

	// check for existing pool
	topoSegment, err = MatchTopologyForPool(&validMultipleTopoPools, &validAccReq, pool1)
	err = checkOutput(err, pool1, topoSegment)
	if err != nil {
		t.Errorf("expected success got: (%v)", err)
	}
}

/*
// TODO: To test GetTopologyFromDomainLabels we need it to accept a k8s client interface, to mock k8sGetNdeLabels output
func TestGetTopologyFromDomainLabels(t *testing.T) {
	fakeNodes := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "worker1",
			Labels: map[string]string{
				"prefix/region": "R1",
				"prefix/zone":   "Z1",
			},
		},
	}

	client := fake.NewSimpleClientset(&fakeNodes)

	_, err := k8sGetNodeLabels(client, "nodeName")
	if err == nil {
		t.Error("Expected error due to invalid node name, got success")
	}

	labels, err := k8sGetNodeLabels(client, "worker1")
	if err != nil {
		t.Errorf("Expected success, got err (%v)", err)
	}
	t.Errorf("Read labels (%v)", labels)
}*/