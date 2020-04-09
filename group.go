package gws

import (
	"errors"
	"sync"
)

type groupManager struct {
	sync.Mutex
	groupCapacity    int
	groupReadyMap    map[string]*group // 满员
	groupNotReadyMap map[string]*group // 未满员
	buildGroupID     func() string     // 生成group id
	Hook
}

// group manager

// 分组
type group struct {
	ID         string
	CtxMap     map[string]*Context
	gm         *groupManager
	cap        int
	currentCap int
	sync.Mutex
}

func NewGM(cap int) *groupManager {
	return &groupManager{
		Mutex:            sync.Mutex{},
		groupCapacity:    cap,
		groupReadyMap:    map[string]*group{},
		groupNotReadyMap: map[string]*group{},
		buildGroupID:     UUID,
	}
}

// 设置生成id func
func (gm *groupManager) SetBuildGroupID(buildGroupID func() string) {
	gm.buildGroupID = buildGroupID
}

// 设置加入group成功之后的hook
func (gm *groupManager) SetJoinGroupHook(afterJoinGroup ...func(c *Context, g *group)) {
	gm.AfterJoinGroup = append(gm.AfterJoinGroup, afterJoinGroup...)
}

func (gm *groupManager) SetLeaveGroupHook(leaveJoinGroup ...func(c *Context, g *group)) {
	gm.AfterJoinGroup = append(gm.AfterJoinGroup, leaveJoinGroup...)
}

// 获取一个未满员组，没有的话创建新组
func (gm *groupManager) GetGroup() *group {
	return gm.newGroup()
}

var ErrorReadyGroup = errors.New("group already exceed limit")
var ErrorGroupNotExist = errors.New("group not exist")

// 创建组
func (gm *groupManager) newGroup() *group {
	g := &group{
		ID:     gm.buildGroupID(),
		CtxMap: map[string]*Context{},
		gm:     gm,
		cap:    gm.groupCapacity,
		Mutex:  sync.Mutex{},
	}
	gm.groupNotReadyMap[g.ID] = g
	return g
}

// 加入组
func (gm *groupManager) JoinGroup(c *Context) {
	gm.Lock()
	defer gm.Unlock()
	var g *group
	if len(gm.groupNotReadyMap) > 0 {
		for _, val := range gm.groupNotReadyMap {
			g = val
		}
	}
	if g == nil {
		g = gm.newGroup()
	}
	g.joinGroup(c)
}

// 加入组
func (g *group) joinGroup(c *Context) {
	g.Lock()
	defer g.Unlock()
	g.CtxMap[c.ID] = c
	c.Group = g
	g.currentCap++
	if g.currentCap == g.currentCap {
		delete(g.gm.groupNotReadyMap, g.ID)
		g.gm.groupReadyMap[g.ID] = g
	}
	// 加入组成功 hook
	for _, f := range g.gm.AfterJoinGroup {
		f(c, g)
	}
}

// 加入指定组
func (gm *groupManager) JoinAGroup(id string, c *Context) error {
	gm.Lock()
	defer gm.Unlock()
	g, ok := gm.groupNotReadyMap[id]
	if ok {
		g.joinGroup(c)
		return nil
	}
	if _, ok := gm.groupReadyMap[id]; ok {
		return ErrorReadyGroup
	}
	return ErrorGroupNotExist
}

func (g *group) LeaveGroup(c *Context) {
	g.Lock()
	defer g.Unlock()
	if _, ok := g.CtxMap[c.ID]; ok {
		delete(g.CtxMap, c.ID)
	}
	// 离开组hook
	for _, f := range g.gm.AfterLeaveGroup {
		f(c, g)
	}
}

func (g *group) SendAll(data interface{}) {
	g.Lock()
	defer g.Unlock()
	for _, c := range g.CtxMap {
		_ = c.SendText(data)
	}
}

func (g *group) SendWithoutC(withC *Context, data interface{}) {
	g.Lock()
	defer g.Unlock()
	for _, c := range g.CtxMap {
		if c.ID == withC.ID {
			continue
		}
		_ = c.SendText(data)
	}
}