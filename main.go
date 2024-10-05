package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	BackpackName  = "рюкзак"
	NoSuchItemMsg = "нет такого"
)

var game *Game

func main() {
	initGame()
	fmt.Println("Добро пожаловать в игру! Введите команду.")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Ошибка чтения команды.")
			break
		}

		command = strings.TrimSpace(command)

		if command == "выход" {
			fmt.Println("Вы вышли из игры.")
			break
		}

		result := handleCommand(command)
		fmt.Println(result)
	}
}

// Game представляет состояние игры
type Game struct {
	Player      *Player
	Rooms       map[string]*Room
	DoorIsOpen  bool
	CurrentRoom *Room
}

// Player представляет состояние игрока
type Player struct {
	Inventory         map[string]*Item
	IsWearingBackpack bool
}

// Room представляет комнату в игре
type Room struct {
	Name    string
	Items   []*Item
	Exits   map[string]*Room
	OnEnter func() string
}

// Item представляет предмет в игре
type Item struct {
	Name     string
	Location string
}

// Инициализация игры
func initGame() {
	game = &Game{
		Player: &Player{
			Inventory:         make(map[string]*Item),
			IsWearingBackpack: false,
		},
		Rooms:      make(map[string]*Room),
		DoorIsOpen: false,
	}

	// Создаем предметы с указанием их местоположения
	tea := &Item{Name: "чай", Location: "на столе"}
	keys := &Item{Name: "ключи", Location: "на столе"}
	notes := &Item{Name: "конспекты", Location: "на столе"}
	backpack := &Item{Name: BackpackName, Location: "на стуле"}

	// Создаем комнаты
	kitchen := &Room{
		Name: "кухня",
		OnEnter: func() string {
			return "кухня, ничего интересного. можно пройти - коридор"
		},
		Items: []*Item{tea},
		Exits: make(map[string]*Room),
	}

	corridor := &Room{
		Name: "коридор",
		OnEnter: func() string {
			return "ничего интересного. можно пройти - кухня, комната, улица"
		},
		Items: []*Item{},
		Exits: make(map[string]*Room),
	}

	room := &Room{
		Name: "комната",
		OnEnter: func() string {
			return "ты в своей комнате. можно пройти - коридор"
		},
		Items: []*Item{keys, notes, backpack},
		Exits: make(map[string]*Room),
	}

	street := &Room{
		Name: "улица",
		OnEnter: func() string {
			return "на улице весна. можно пройти - домой"
		},
		Items: []*Item{},
		Exits: make(map[string]*Room),
	}

	// Соединяем комнаты между собой
	kitchen.Exits["коридор"] = corridor
	corridor.Exits["кухня"] = kitchen
	corridor.Exits["комната"] = room
	corridor.Exits["улица"] = street
	room.Exits["коридор"] = corridor
	street.Exits["домой"] = corridor

	// Задаём начальную комнату
	game.Rooms["кухня"] = kitchen
	game.Rooms["коридор"] = corridor
	game.Rooms["комната"] = room
	game.Rooms["улица"] = street
	game.CurrentRoom = kitchen
}

// Обработка команд от игрока
func handleCommand(command string) string {
	// Разбиваем команду на слова
	words := strings.Split(command, " ")
	action := words[0]
	args := words[1:]

	switch action {
	case "осмотреться":
		return game.Player.lookAround()
	case "идти":
		if len(args) == 0 {
			return "куда идти?"
		}
		return game.Player.moveTo(strings.Join(args, " "))
	case "взять":
		if len(args) == 0 {
			return "что взять?"
		}
		return game.Player.takeItem(strings.Join(args, " "))
	case "надеть":
		if len(args) == 0 {
			return "что надеть?"
		}
		return game.Player.wearItem(strings.Join(args, " "))
	case "применить":
		if len(args) < 2 {
			return "не к чему применить"
		}
		return game.Player.useItem(args[0], args[1])
	default:
		return "неизвестная команда"
	}
}

// Функция осмотра комнаты
func (p *Player) lookAround() string {
	room := game.CurrentRoom

	switch room.Name {
	case "кухня":
		if p.IsWearingBackpack {
			return "ты находишься на кухне, на столе: чай, надо идти в универ. можно пройти - коридор"
		} else {
			return "ты находишься на кухне, на столе: чай, надо собрать рюкзак и идти в универ. можно пройти - коридор"
		}
	case "комната":
		var itemsAtTable []string
		var itemsAtChair []string
		for _, item := range room.Items {
			if item.Name == BackpackName && p.IsWearingBackpack {
				continue // Рюкзак уже надет, его нет в комнате
			}
			if item.Location == "на столе" {
				itemsAtTable = append(itemsAtTable, item.Name)
			} else if item.Location == "на стуле" {
				itemsAtChair = append(itemsAtChair, item.Name)
			}
		}
		if len(itemsAtTable) == 0 && len(itemsAtChair) == 0 {
			return "пустая комната. можно пройти - коридор"
		}
		var descParts []string
		if len(itemsAtTable) > 0 {
			descParts = append(descParts, fmt.Sprintf("на столе: %s", strings.Join(itemsAtTable, ", ")))
		}
		if len(itemsAtChair) > 0 {
			descParts = append(descParts, fmt.Sprintf("на стуле: %s", strings.Join(itemsAtChair, ", ")))
		}
		return strings.Join([]string{strings.Join(descParts, ", "), "можно пройти - коридор"}, ". ")
	default:
		return room.OnEnter()
	}
}

// Функция перемещения между комнатами
func (p *Player) moveTo(destination string) string {
	if destination == "улица" && !game.DoorIsOpen {
		return "дверь закрыта"
	}

	if nextRoom, ok := game.CurrentRoom.Exits[destination]; ok {
		game.CurrentRoom = nextRoom
		return nextRoom.OnEnter()
	}

	return strings.Join([]string{"нет пути в", destination}, " ")
}

// Функция для взятия предметов
func (p *Player) takeItem(itemName string) string {
	if !p.IsWearingBackpack {
		return "некуда класть"
	}
	room := game.CurrentRoom
	for i, item := range room.Items {
		if item.Name == itemName {
			p.Inventory[itemName] = item
			// Удаляем предмет из комнаты
			room.Items = append(room.Items[:i], room.Items[i+1:]...)
			return strings.Join([]string{"предмет добавлен в инвентарь:", itemName}, " ")
		}
	}
	return NoSuchItemMsg
}

// Функция для надевания предметов
func (p *Player) wearItem(itemName string) string {
	if itemName == BackpackName && game.CurrentRoom.Name == "комната" {
		for i, item := range game.CurrentRoom.Items {
			if item.Name == BackpackName {
				p.IsWearingBackpack = true
				// Удаляем предмет из комнаты
				game.CurrentRoom.Items = append(game.CurrentRoom.Items[:i], game.CurrentRoom.Items[i+1:]...)
				return "вы надели: рюкзак"
			}
		}
		return NoSuchItemMsg
	}
	return NoSuchItemMsg
}

// Функция для применения предметов
func (p *Player) useItem(itemName, target string) string {
	if _, ok := p.Inventory[itemName]; !ok {
		return strings.Join([]string{"нет предмета в инвентаре -", itemName}, " ")
	}

	if itemName == "ключи" && target == "дверь" && game.CurrentRoom.Name == "коридор" {
		game.DoorIsOpen = true
		return "дверь открыта"
	}

	return "не к чему применить"
}
