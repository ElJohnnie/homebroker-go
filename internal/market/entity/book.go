package entity

import (
	"container/heap"
	"sync"
)

type Book struct {
	Order []*Order
	Transaction []*Transaction
	OrdersChannel chan *Order
	OrdersChannelOut chan *Order
	WaitGroup *sync.WaitGroup
}

func NewBook(OrdersChannel chan *Order, OrdersChannelOut chan *Order, waitGroup *sync.WaitGroup) *Book {
	return &Book{
		Order: []*Order{},
		Transaction: []*Transaction{},
		OrdersChannel: OrdersChannel,
		OrdersChannelOut: OrdersChannelOut,
		WaitGroup: waitGroup,
	}
}

func (b *Book) Trade() {
	buyOrders := NewOrderQueue()
	sellOrders := NewOrderQueue()

	heap.Init(buyOrders)
	heap.Init(sellOrders)

	for order := range b.OrdersChannel {
		if order.OrderType == "BUY" {
			buyOrders.Push(order)
			if sellOrders.Len() > 0 && sellOrders.Orders[0].Price <= order.Price {
				sellOrder := sellOrders.Pop().(*Order)
				if sellOrder.PendingShares > 0 {
					transaction := NewTransaction(sellOrder, order, order.Shares, sellOrder.Price)
					b.AddTransaction(transaction, b.WaitGroup)
					sellOrder.Transactions = append(sellOrder.Transactions, transaction)
					order.Transactions = append(order.Transactions, transaction)
					b.OrdersChannelOut <- sellOrder
					b.OrdersChannel <- order
					if sellOrder.PendingShares > 0 {
						sellOrders.Push(sellOrder)
					}
				}
			}
		} else if order.OrderType == "SELL" {
			sellOrders.Push(order)
			if buyOrders.Len() > 0 && buyOrders.Orders[0].Price <= order.Price {
				buyOrder := buyOrders.Pop().(*Order)
				if buyOrder.PendingShares > 0 {
					transaction := NewTransaction(buyOrder, order, order.Shares, buyOrder.Price)
					b.AddTransaction(transaction, b.WaitGroup)
					buyOrder.Transactions = append(buyOrder.Transactions, transaction)
					order.Transactions = append(order.Transactions, transaction)
					b.OrdersChannelOut <- buyOrder
					b.OrdersChannel <- order
					if buyOrder.PendingShares > 0 {
						buyOrders.Push(buyOrder)
					}
				}
			}
		}
	}
}

func (b *Book) AddTransaction(transaction *Transaction, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	sellingShares := transaction.SellingOrder.PendingShares
	buyingShares := transaction.BuyingOrder.PendingShares

	minShares := sellingShares
	if buyingShares < minShares {
		minShares = buyingShares
	}

	transaction.SellingOrder.Investor.UpdateAssetPosition(transaction.SellingOrder.Asset.ID, -minShares)
	transaction.SellingOrder.PendingShares -= minShares
	transaction.BuyingOrder.Investor.UpdateAssetPosition(transaction.SellingOrder.Asset.ID, minShares)
	transaction.BuyingOrder.PendingShares -= minShares

	transaction.Total = transaction.BuyingOrder.Price * float64(transaction.Price)

	if transaction.SellingOrder.PendingShares == 0 {
		transaction.SellingOrder.Status = "CLOSED"
	}

	if transaction.BuyingOrder.PendingShares == 0 {
		transaction.BuyingOrder.Status = "CLOSED"
	}

	b.Transaction = append(b.Transaction, transaction)
}