/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package authcenter

import (
	"errors"
	"fmt"

	"configcenter/src/auth/meta"
)

// Adaptor is a middleware wrapper which works for converting concepts
// between bk-cmdb and blueking auth center. Especially the policies
// in auth center.

func adaptor(attribute *meta.ResourceAttribute) (*ResourceInfo, error) {
	resourceType := attribute.Basic.Type
	info := new(ResourceInfo)
	info.ResourceName = attribute.Basic.Name

	var err error
	info.ResourceID, err = GenerateResourceID(attribute)
	if err != nil {
		return nil, err
	}

	switch resourceType {
	case meta.Business:
		info.ResourceType = SysBusinessInstance
		return info, nil

	case meta.Model,
		meta.ModelUnique,
		meta.ModelAttribute,
		meta.ModelAttributeGroup:
		if attribute.BusinessID != 0 {
			info.ResourceType = BizModel
		} else {
			info.ResourceType = SysModel
		}

	case meta.ModelModule, meta.ModelSet, meta.ModelInstanceTopology:
		info.ResourceType = BizTopoInstance

	case meta.MainlineModel, meta.ModelTopology:
		info.ResourceType = SysSystemBase

	case meta.ModelClassification:
		info.ResourceType = SysModelGroup

	case meta.AssociationType:
		info.ResourceType = SysAssociationType

	case meta.ModelAssociation:
		return info, errors.New("model association does not support auth now")

	case meta.ModelInstanceAssociation:
		return info, errors.New("model instance association does not support  auth now")

	case meta.ModelInstance:
		if attribute.BusinessID == 0 {
			info.ResourceType = SysInstance
		} else {
			info.ResourceType = BizInstance
		}

	case meta.HostInstance:
		if attribute.BusinessID == 0 {
			info.ResourceType = SysHostInstance
		} else {
			info.ResourceType = BizHostInstance
		}

	case meta.HostUserCustom:
		info.ResourceType = BizCustomQuery

	case meta.HostFavorite:
		return info, errors.New("host favorite does not support auth now")

	case meta.Process:
		info.ResourceType = BizProcessInstance

	case meta.NetDataCollector:
		return nil, fmt.Errorf("unsupported resource type: %s", attribute.Basic.Type)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", attribute.Basic.Type)
	}

	return info, nil
}

// ResourceTypeID is resource's type in auth center.
type ResourceTypeID string

// System Resource
const (
	SysSystemBase       ResourceTypeID = "sysSystemBase"
	SysBusinessInstance ResourceTypeID = "sysBusinessInstance"
	SysHostInstance     ResourceTypeID = "sysHostInstance"
	SysEventPushing     ResourceTypeID = "sysEventPushing"
	SysModelGroup       ResourceTypeID = "sysModelGroup"
	SysModel            ResourceTypeID = "sysModel"
	SysInstance         ResourceTypeID = "sysInstance"
	SysAssociationType  ResourceTypeID = "sysAssociationType "
)

// Business Resource
const (
	// the alias name maybe "dynamic classification"
	BizCustomQuery     ResourceTypeID = "bizCustomQuery"
	BizHostInstance    ResourceTypeID = "bizHostInstance"
	BizProcessInstance ResourceTypeID = "bizProcessInstance"
	BizTopoInstance    ResourceTypeID = "bizTopoInstance"
	BizModelGroup      ResourceTypeID = "bizModelGroup"
	BizModel           ResourceTypeID = "bizModel"
	BizInstance        ResourceTypeID = "bizInstance"
)

type ActionID string

// ActionID define
const (
	// Unknown action is a action that can not be recognized by the auth center.
	Unknown ActionID = "unknown"
	Edit    ActionID = "edit"
	Create  ActionID = "create"
	Get     ActionID = "get"
	Delete  ActionID = "delete"

	// Archive for business
	Archive ActionID = "archive"
	// host action
	ModuleTransfer ActionID = "moduleTransfer"
	// business topology action
	HostTransfer ActionID = "hostTransfer"
	// system base action, related to model topology
	ModelTopologyView ActionID = "modelTopologyView"
	// business model topology operation.
	ModelTopologyOperation ActionID = "modelTopologyOperation"
	// assign host(s) to a business
	// located system/host/assignHostsToBusiness in auth center.
	AssignHostsToBusiness ActionID = "assignHostsToBusiness"
	BindModule            ActionID = "bindModule"
	BindModuleQuery       ActionID = "bindModuleQuery"
	AdminEntrance         ActionID = "adminEntrance"
)

func adaptorAction(r *meta.ResourceAttribute) (ActionID, error) {

	if r.Action == meta.Find || r.Action == meta.Delete || r.Action == meta.Create {
		if r.Basic.Type == meta.MainlineModel {
			return ModelTopologyOperation, nil
		}
	}

	if r.Action == meta.Find || r.Action == meta.Update {
		if r.Basic.Type == meta.ModelTopology {
			return ModelTopologyView, nil
		}
	}

	if r.Basic.Type == meta.Process {
		if r.Action == meta.BoundModuleToProcess || r.Action == meta.UnboundModuleToProcess {
			return BindModule, nil
		}

		if r.Action == meta.FindBoundModuleProcess {
			return BindModuleQuery, nil
		}
	}

	switch r.Action {
	case meta.Create, meta.CreateMany:
		return Create, nil

	case meta.Find, meta.FindMany:
		return Get, nil

	case meta.Delete, meta.DeleteMany:
		return Delete, nil

	case meta.Update, meta.UpdateMany:
		return Edit, nil

	case meta.MoveResPoolHostToBizIdleModule:
		if r.Basic.Type == meta.ModelInstance && r.Basic.Name == meta.Host {
			return AssignHostsToBusiness, nil
		}

	case meta.MoveHostToBizFaultModule,
		meta.MoveHostToBizIdleModule,
		meta.MoveHostFromModuleToResPool,
		meta.MoveHostToAnotherBizModule,
		meta.CleanHostInSetOrModule,
		meta.MoveHostToModule:
		if r.Basic.Type == meta.ModelInstance && r.Basic.Name == meta.Host {
			return ModuleTransfer, nil
		}

	case meta.AddHostToResourcePool:
		// add hosts to resource pool
		if r.Basic.Type == meta.ModelInstance && r.Basic.Name == meta.Host {
			return Create, nil
		}
		return ModuleTransfer, nil

	case meta.MoveHostsToBusinessOrModule:
		return Edit, nil

	}

	return Unknown, fmt.Errorf("unsupported action: %s", r.Action)
}

type ResourceDetail struct {
	// the resource type in auth center.
	Type ResourceTypeID
	// all the actions that this resource supported.
	Actions []ActionID
}

var (
	CustomQueryDescribe = ResourceDetail{
		Type:    BizCustomQuery,
		Actions: []ActionID{Get, Delete, Edit, Create},
	}

	AppModelDescribe = ResourceDetail{
		Type:    BizModel,
		Actions: []ActionID{Get, Delete, Edit, Create},
	}

	HostDescribe = ResourceDetail{
		Type:    BizHostInstance,
		Actions: []ActionID{Get, Delete, Edit, Create, ModuleTransfer},
	}

	ProcessDescribe = ResourceDetail{
		Type:    BizProcessInstance,
		Actions: []ActionID{Get, Delete, Edit, Create},
	}

	TopologyDescribe = ResourceDetail{
		Type:    BizTopoInstance,
		Actions: []ActionID{Get, Delete, Edit, Create, HostTransfer},
	}

	AppInstanceDescribe = ResourceDetail{
		Type:    BizInstance,
		Actions: []ActionID{Get, Delete, Edit, Create},
	}

	InstanceManagementDescribe = ResourceDetail{
		Type:    SysInstance,
		Actions: []ActionID{Get, Delete, Edit, Create},
	}

	ModelManagementDescribe = ResourceDetail{
		Type:    SysModel,
		Actions: []ActionID{Get, Delete, Edit, Create},
	}

	AssociationTypeDescribe = ResourceDetail{
		Type:    SysAssociationType,
		Actions: []ActionID{Get, Delete, Edit, Create},
	}

	ModelGroupDescribe = ResourceDetail{
		Type:    SysModelGroup,
		Actions: []ActionID{Get, Delete, Edit, Create},
	}

	EventDescribe = ResourceDetail{
		Type:    SysEventPushing,
		Actions: []ActionID{Get, Delete, Edit, Create},
	}

	SystemBaseDescribe = ResourceDetail{
		Type:    SysSystemBase,
		Actions: []ActionID{Get, Delete, Edit, Create},
	}
)
