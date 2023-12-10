package gnet

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

type Pool struct {
	idlesConns  chan *idleConn // 空闲连接队列
	reqQueue    []connReq      // 	请求队列
	maxCnt      int            // 最大连接数
	cnt         int            // 当前连接数
	maxIdleTime time.Duration  // 最大空闲连接
	factory     func() (net.Conn, error)
	lock        sync.Mutex
}

func NewPool(initCnt int, maxIdleCnt int, maxCnt int,
	maxIdleTime time.Duration, factory func() (net.Conn, error)) (*Pool, error) {
	if initCnt > maxIdleCnt {
		return nil, errors.New("gnet: 初始连接数量不能大于最大空闲连接数量")
	}
	idlesConns := make(chan *idleConn, maxIdleCnt)
	for i := 0; i < initCnt; i++ {
		conn, err := factory()
		if err != nil {
			return nil, err
		}
		idlesConns <- &idleConn{c: conn, lastActivateTime: time.Now()} // 加入空闲连接
	}
	res := &Pool{
		idlesConns:  idlesConns,
		maxCnt:      maxCnt,
		cnt:         0,
		maxIdleTime: maxIdleTime,
		factory:     factory,
	}
	return res, nil
}

// Get 获取连接
func (p *Pool) Get(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:

	}
	for {
		select {
		case ic := <-p.idlesConns: // 拿去空闲连接
			if ic.lastActivateTime.Add(p.maxIdleTime).Before(time.Now()) { // 过期校验
				_ = ic.c.Close() // 过期关闭
				continue
			}
			return ic.c, nil
		default:
			// 没有空闲连接
			p.lock.Lock()
			if p.cnt >= p.maxCnt {
				// 超过上限
				req := connReq{
					connChan: make(chan net.Conn, 1),
				}
				p.reqQueue = append(p.reqQueue, req) // 加入阻塞队列
				p.lock.Unlock()
				select { // 阻塞等待
				case <-ctx.Done(): // 被取消了
					// 方式1：从队列里面删除req
					// 方式2：在这里进行转发
					go func() {
						c := <-req.connChan                // 别人归还
						_ = p.Put(context.Background(), c) // 转发放入
					}()
					return nil, ctx.Err()
				case c := <-req.connChan: // 等别人归还
					return c, nil
				}
			}
			// 没有超过上限
			c, err := p.factory() // 创建连接
			if err != nil {
				return nil, err
			}
			p.cnt++
			p.lock.Unlock()
			return c, nil
		}
	}
}
func (p *Pool) Put(ctx context.Context, c net.Conn) error {
	p.lock.Lock()
	if len(p.reqQueue) > 0 {
		// 有阻塞的请求
		req := p.reqQueue[0] // 唤醒请求
		p.reqQueue = p.reqQueue[1:]
		p.lock.Unlock()
		req.connChan <- c // 转交连接
		return nil
	}
	defer p.lock.Unlock()
	// 没有阻塞的请求
	ic := &idleConn{
		c:                c,
		lastActivateTime: time.Now(),
	}
	select {
	case p.idlesConns <- ic: // 放回队列
	default:
		// 空闲队列满了
		_ = c.Close() // 关闭连接
		p.cnt--
	}
	return nil
}

type idleConn struct {
	c                net.Conn
	lastActivateTime time.Time
}
type connReq struct {
	connChan chan net.Conn
}
