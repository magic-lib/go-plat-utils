package state

import (
	"fmt"
	"github.com/samber/lo"
)

// BaseFsm 状态机结构体
type BaseFsm[S comparable, A comparable] struct {
	// 状态映射关系图
	// <State, ActionMap>
	// ActionMap --> <Action, NextState>
	stateMap map[S]map[A]S // 状态映射表，key为当前状态，value为动作到目标状态的映射
	// 记录前状态
	preState S
	// 记录当前状态
	currState S
	// allStates 避免后续添加一些不支持的状态和动作，前期统一注册完毕后，不允许再添加
	allStates  []S
	allActions []A
}

// NewBaseFsm 构造方法
func NewBaseFsm[S comparable, A comparable](currState ...S) *BaseFsm[S, A] {
	one := &BaseFsm[S, A]{
		stateMap: make(map[S]map[A]S),
	}
	if len(currState) >= 1 {
		one.preState = currState[0]
		one.currState = currState[0]
	}
	return one
}

// Transition 执行状态转换
func (bf *BaseFsm[S, A]) Transition(action A) (bool, error) {
	nextState, err := bf.NextState(bf.currState, action)
	if err != nil {
		return false, err
	}
	bf.preState = bf.currState
	bf.currState = nextState
	return true, nil
}
func (bf *BaseFsm[S, A]) CanTransition(currState S, action A, toState S) bool {
	nextState, err := bf.NextState(currState, action)
	if err != nil {
		return false
	}
	if nextState == toState {
		return true
	}
	return false
}

func (bf *BaseFsm[S, A]) checkState(state S) bool {
	if bf.allStates == nil || len(bf.allStates) == 0 {
		return true
	}
	return lo.Contains(bf.allStates, state)
}
func (bf *BaseFsm[S, A]) checkAction(action A) bool {
	if bf.allActions == nil || len(bf.allActions) == 0 {
		return true
	}
	return lo.Contains(bf.allActions, action)
}

// Register 注册状态变迁关系
func (bf *BaseFsm[S, A]) Register(stateFrom S, action A, stateTo S) error {
	if !bf.checkState(stateFrom) || !bf.checkAction(action) || !bf.checkState(stateTo) {
		return fmt.Errorf("stateFrom: %v, action: %v, stateTo: %v not exists", stateFrom, action, stateTo)
	}

	actMap, exists := bf.stateMap[stateFrom]
	if !exists {
		actMap = make(map[A]S)
		bf.stateMap[stateFrom] = actMap
	}
	if _, exists := actMap[action]; exists {
		return fmt.Errorf("stateFrom: %v, action: %v, stateTo: %v already exists", stateFrom, action, stateTo)
	}
	actMap[action] = stateTo
	return nil
}

// RegisterAll 注册所有状态和动作，避免后期添加一些不支持的状态和动作
func (bf *BaseFsm[S, A]) RegisterAll(state []S, action []A) {
	bf.allStates = state
	bf.allActions = action
}

// CurrState 获取当前状态
func (bf *BaseFsm[S, A]) CurrState() S {
	return bf.currState
}

// SetCurrState 设置当前状态
func (bf *BaseFsm[S, A]) SetCurrState(currState S) {
	bf.currState = currState
}

// PreState 获取前一状态
func (bf *BaseFsm[S, A]) PreState() S {
	return bf.preState
}

// StateMap 获取所有状态变迁关系
func (bf *BaseFsm[S, A]) StateMap() map[S]map[A]S {
	return bf.stateMap
}

// NextState 检查从当前状态通过某个动作是否可以转换到目标状态
func (bf *BaseFsm[S, A]) NextState(currState S, action A) (s S, err error) {
	if !bf.checkState(currState) || !bf.checkAction(action) {
		return s, fmt.Errorf("currState: %v, action: %v, not exists", currState, action)
	}
	if actMap, exists := bf.stateMap[currState]; exists {
		if stateTo, ok := actMap[action]; ok {
			return stateTo, nil
		}
	}
	return s, fmt.Errorf("no next state found for currState: %v, action: %v", currState, action)
}

// NextActionList 返回可执行的动作列表
func (bf *BaseFsm[S, A]) NextActionList(currState S) []A {
	if actMap, exists := bf.stateMap[currState]; exists {
		return lo.Keys(actMap)
	}
	return []A{}
}
