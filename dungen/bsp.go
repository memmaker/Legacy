package dungen

import (
    "Legacy/geometry"
    "math/rand"
)

type DungeonGenerator interface {
    Generate(width, height int) *DungeonMap
}

type BSPNode struct {
    rect       geometry.Rect
    parent     *BSPNode
    leftChild  *BSPNode
    rightChild *BSPNode
    sibling    *BSPNode

    containedRoom      *DungeonRoom
    containedCorridors []geometry.Rect
}

func (n *BSPNode) IsLeaf() bool {
    return n.leftChild == nil && n.rightChild == nil
}

func (n *BSPNode) TraverseLeafs(f func(node *BSPNode)) {
    if n.IsLeaf() {
        f(n)
        return
    } else {
        if n.leftChild != nil {
            n.leftChild.TraverseLeafs(f)
        }
        if n.rightChild != nil {
            n.rightChild.TraverseLeafs(f)
        }
    }
}

func (n *BSPNode) TraverseSiblingPairs(f func(node *BSPNode, sibling *BSPNode)) {
    if n.sibling != nil {
        f(n, n.sibling)
    }
    if n.leftChild != nil {
        n.leftChild.TraverseSiblingPairs(f)
    }
    if n.rightChild != nil {
        n.rightChild.TraverseSiblingPairs(f)
    }
}

func (n *BSPNode) GetRandomLeaf() *BSPNode {
    if n.IsLeaf() {
        return n
    } else {
        if 0.5 < rand.Float64() {
            return n.leftChild.GetRandomLeaf()
        } else {
            return n.rightChild.GetRandomLeaf()
        }
    }
}

type BSPGenerator struct {
}

func NewBSPGenerator() *BSPGenerator {
    return &BSPGenerator{}
}

func (g *BSPGenerator) Generate(width, height int) *DungeonMap {
    dMap := NewDungeonMap(width, height)

    return dMap
}

func (g *BSPGenerator) getBSPRootNode(width int, height int) *BSPNode {
    currentRect := geometry.NewRect(0, 0, width, height)
    rootNode := g.subdivide(&BSPNode{
        rect: currentRect,
    })
    return rootNode
}

func (g *BSPGenerator) subdivide(currentNode *BSPNode) *BSPNode {
    minSize := 4

    var partOne, partTwo geometry.Rect
    currentRect := currentNode.rect
    // chose axis
    if currentRect.Size().X > currentRect.Size().Y {
        minSpaceNeeded := 2 * minSize
        if currentRect.Size().X <= minSpaceNeeded {
            return currentNode
        }
        spaceInterval := currentRect.Size().X - minSpaceNeeded
        randomX := minSize + rand.Intn(spaceInterval)
        partOne, partTwo = currentRect.BisectAtColumn(randomX)
        nodeOne := g.subdivide(&BSPNode{
            parent: currentNode,
            rect:   partOne,
        })
        currentNode.leftChild = nodeOne
        nodeTwo := g.subdivide(&BSPNode{
            parent: currentNode,
            rect:   partTwo,
        })
        currentNode.rightChild = nodeTwo
        nodeOne.sibling = nodeTwo
        nodeTwo.sibling = nodeOne
    } else {
        minSpaceNeeded := 2 * minSize
        if currentRect.Size().Y <= minSpaceNeeded {
            return currentNode
        }
        spaceInterval := currentRect.Size().Y - minSpaceNeeded
        randomY := minSize + rand.Intn(spaceInterval)
        partOne, partTwo = currentRect.BisectAtLine(randomY)

        nodeOne := g.subdivide(&BSPNode{
            parent: currentNode,
            rect:   partOne,
        })
        currentNode.leftChild = nodeOne
        nodeTwo := g.subdivide(&BSPNode{
            parent: currentNode,
            rect:   partTwo,
        })
        currentNode.rightChild = nodeTwo
        nodeOne.sibling = nodeTwo
        nodeTwo.sibling = nodeOne
    }
    return currentNode
}
