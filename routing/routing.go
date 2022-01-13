package routing

import (
	// "fmt"

	"fmt"

	"github.com/takoyaki-3/minmaxRouting"
	pb "github.com/takoyaki-3/minmaxRouting/pb"
)

type Query struct {
	NWeight int
	FromNode minmaxrouting.NodeIdType
	ToNode minmaxrouting.NodeIdType
	MaxTransfer int
	IsSerialNG bool
}

type Route struct {
	Nodes []minmaxrouting.NodeIdType
	Weight minmaxrouting.Weight
	SubNodes [][]minmaxrouting.NodeIdType
}

type Memo [][]*CB

func newCB(data *CB)*CB{
	cb := new(CB)
	*cb = *data
	return cb
}

func MinMaxRouting(g *minmaxrouting.Graph,query Query)(routes []Route,memo Memo){

	toNode := minmaxrouting.NodeIdType(query.ToNode)

	// マーティン拡張による最大・最小の
	memo = make(Memo,len(g.Nodes))

	// 初期化
	initEdge := minmaxrouting.EdgeIdType(-1)

	memo[query.FromNode] = append(memo[query.FromNode],newCB(&CB{
		BeforeEdgeId: initEdge,
		Weight: *minmaxrouting.NewWeight(query.NWeight),
		Node: query.FromNode,
		BeforeCB: nil,
	}))

	cou := 0
	que := Que{}
	que.Add(memo[query.FromNode][0])

	for que.Len() > 0 {
		posCB := que.Get()

		cou++
		if toNode >= 0 && cou % 1000 == 0 {
			fmt.Println(cou,que.Len(),len(memo[toNode]))
		}

		// 目的地が存在するクエリの場合、目的地到達判定を行う
		if toNode != -1 {
			// 枝刈り：既知のゴールへの重みより大きいか検証
			if !BetterIndex(posCB.Weight,&memo[toNode]) {
				continue
			}
		}

		// Posが始点となる辺から探索範囲拡張を行っていく
		for _,edgeId := range g.Nodes[posCB.Node].FromEdgeIds{
			edge := g.Edges[edgeId]

			// 同一辺タイプ（コンポーネント毎に異なるEdgeTypeIdを想定）であればスキップ
			if query.IsSerialNG && posCB.BeforeEdgeId != initEdge && edge.EdgeTypeId == g.Edges[posCB.BeforeEdgeId].EdgeTypeId{
				continue
			}
			
			// 乗換辺を連続使用しない。乗換辺はEdgeTypeIdは-1。mkTreeの時に使用する。
			if posCB.BeforeEdgeId != initEdge && edge.EdgeTypeId == -1 && g.Edges[posCB.BeforeEdgeId].EdgeTypeId == -1{
				continue
			}

			// // 同一路線を排除
			// if !query.IsSerialNG {
			// 	if memoList[posCBIndex].BeforeEdgeId != initEdge{
			// 		beforeUseTrips := g.Edges[memoList[posCBIndex].BeforeEdgeId].UseTrips
			// 		if len(beforeUseTrips) > 0{
			// 			befS := -1
			// 			for tripIndex,trip := range edge.UseTrips{
			// 				if trip == beforeUseTrips[0]{
			// 					befS = tripIndex
			// 				}
			// 			}
			// 			if befS != -1{
			// 				flag := false
			// 				for index:=0;index<len(edge.UseTrips)-befS;index++{
			// 					if len(beforeUseTrips) == index{
			// 						flag = true
			// 						break
			// 					}
			// 					if edge.UseTrips[befS+index] != beforeUseTrips[index]{
			// 						flag = true
			// 						break
			// 					}
			// 				}
			// 				if !flag {
			// 					continue
			// 				}
			// 			}
			// 		}
			// 	}
			// }

			newW := WeightAdder(posCB.Weight,edge.Weight)

			// 乗換え回数の判定
			if int(newW.Weights[1].Min) > query.MaxTransfer {
				continue
			}

			// 既知のゴールへの重みより大きいか検証
			if toNode != -1 && !BetterIndex(newW,&memo[toNode]) {
				continue
			}
			f := true
			for index := 0;index < len(memo[edge.ToId]);index++{
				mi := memo[edge.ToId][index]
				if mi == nil {
					memo[edge.ToId] = append(memo[edge.ToId][:index],memo[edge.ToId][index+1:]...)
					index--
					continue
				}
				cw := CompWeight(mi.Weight,newW)
				if cw == -1 {
					f = false
					break
				} else if cw == 1 {
					memo[edge.ToId] = append(memo[edge.ToId][:index],memo[edge.ToId][index+1:]...)
					index--
				}
			}
			if !f{
				continue
			}

			// 枝刈り施策
			flag := true
			// 辺の経由駅が訪問済みの場合
			for _,n := range edge.ViaNodes{
				if edge.ToId == n{
					flag = false
					break
				}
			}
			if !flag{
				continue
			}

			// 既に訪問済みの頂点であればスキップ
			if posCB.BeforeCB != nil{
				bposcb := posCB.BeforeCB
				flag := true
				for {
					if bposcb.BeforeEdgeId == initEdge {
						break
					}
					if bposcb.Node == edge.ToId {
						flag = false
						break
					}
					bposcb = bposcb.BeforeCB
				}
				if !flag {
					continue
				}
			}

			ncb := newCB(&CB{
				Node: edge.ToId,
				BeforeEdgeId: edgeId,
				Weight: newW,
				BeforeCB: posCB,
			})

			que.Add(ncb)
			memo[edge.ToId] = append(memo[edge.ToId], ncb)

			// デバッグ用出力
			// fmt.Print("edgeId:",edgeId,"	",cou,"	edge ",edge.FromId,edge.ToId,g.Nodes[memoList[posCBIndex].Node].FromEdgeIds,"	",edge.ToId)
			// if memoList[posCBIndex].BeforeCB > 0{
			// 	bposcb := posCb
			// 	for {
			// 		fmt.Print(".",bposcb.Node)
			// 		if bposcb.BeforeEdgeId == initEdge {
			// 			break
			// 		}
			// 		if bposcb.Node == edge.ToId {
			// 			break
			// 		}
			// 		bposcb = memoList[bposcb.BeforeCB]
			// 	}
			// }
			// fmt.Println("")
		}
	}

	if toNode != -1 {
		for _,m:=range memo[toNode] {
			// if m != nil{
				if BetterIndex(m.Weight,&memo[toNode]) {
					r := GetRoute(m)
					if r != nil {
						routes = append(routes, *r)
					}
				}
			// }
		}
	}

	return routes,memo
}

func GetRoutes(cbs *[]*CB)[]Route{
	routes := []Route{}
	for _,cb := range *cbs{
		if Better(cb.Weight,cbs) {
			r := GetRoute(cb.BeforeCB)
			if r != nil{
				routes = append(routes, *r)
			}
		}
	}
	return routes
}

func GetRouteTree(memo Memo)(*pb.RouteTree){

	memoList := []CB{}
	cbmap := map[*CB]int{}
	cbmap[nil] = -3
	for _,v:=range memo{
		for _,p:=range v{
			if _,ok:=cbmap[p];!ok{
				cbmap[p] = len(memoList)
				memoList = append(memoList, *p)
			}	
		}
	}

	tree := &pb.RouteTree{}
	tree.Leaves = make([]*pb.Leaf, len(memoList))
	for index,cb:=range memoList{
		wight := &pb.Weight{}
		for _,w := range cb.Weight.Weights {
			wight.Max = append(wight.Max, int32(w.Max))
			wight.Min = append(wight.Min, int32(w.Min))
		}
		tree.Leaves[index] = &pb.Leaf{
			IsUse: true,
			NodeId: int32(cb.Node),
			Index: int32(cbmap[cb.BeforeCB]),
			BeforeEdgeId: int32(cb.BeforeEdgeId),
			Weight: wight,
		}
	}
	return tree
}

func GetRoute(pos *CB)*Route{
	route := Route{
		Weight: pos.Weight,
	}
	for pos != nil {
		route.Nodes = append([]minmaxrouting.NodeIdType{pos.Node},route.Nodes...)
		pos = pos.BeforeCB
	}
	return &route
}

func Better(weight minmaxrouting.Weight,ws *[]*CB)bool{
	for _,w := range *ws{
		if w != nil{
			if CompWeight(weight,w.Weight) == 1 {
				return false
			}	
		}
	}
	return true
}
func BetterIndex(weight minmaxrouting.Weight,ws *[]*CB)bool{
	for _,w := range *ws{
		if w != nil {
			if len(w.Weight.Weights) > 0{
				if CompWeight(weight,w.Weight) == 1 {
					return false
				}	
			}
		}
	}
	return true
}

func WeightAdder(w1 minmaxrouting.Weight,w2 minmaxrouting.Weight)(w minmaxrouting.Weight){
	for i,_ := range w1.Weights{
		w.Weights = append(w.Weights, minmaxrouting.MinMax{
			Min: w1.Weights[i].Min+w2.Weights[i].Min,
			Max: w1.Weights[i].Max+w2.Weights[i].Max,
		})
	}
	return w
}

type Que struct {
	Cb []*CB
}

func (q *Que)Add(p *CB){
	q.Cb = append(q.Cb, p)
}

func (q *Que)Len()int{
	return len(q.Cb)
}

func (q *Que)Get()(*CB){
	cb := q.Cb[0]
	q.Cb = q.Cb[1:]
	return cb
}

// w1が良い場合-1、w2が良い場合1、一概に決まらない場合0
func CompWeight(w1 minmaxrouting.Weight,w2 minmaxrouting.Weight)int{
	is1Better := true
	is2Better := true

	for i,_:=range w1.Weights{
		if !(w1.Weights[i].Max <= w2.Weights[i].Min) {
			// w1が一概に優っているとは言えない
			is1Better = false
		}
		if !(w2.Weights[i].Max <= w1.Weights[i].Min) {
			// w1が一概に優っているとは言えない
			is2Better = false
		}
	}
	if is1Better == is2Better {
		return 0
	}
	if is1Better {
		return -1
	}
	return 1
}

type CB struct {
	Node minmaxrouting.NodeIdType
	BeforeEdgeId minmaxrouting.EdgeIdType
	Weight minmaxrouting.Weight
	BeforeCB *CB
}

