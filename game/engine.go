package game

import (
    "image/color"
)

type Engine interface {
    StartConversation(a *Actor)
    ShowScrollableText(text []string, textcolor color.Color)
    GetTextFile(filename string) []string
    PickUpItem(item Item)
    DropItem(item Item)
    GetAvatar() *Actor
    IsPlayerControlled(holder ItemHolder) bool
    SwitchAvatarTo(member *Actor)
}
