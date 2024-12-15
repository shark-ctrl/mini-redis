package main

import (
	"math/rand"
)

func zslCreate() *zskiplist {
	var j int
	//初始化跳表结构体
	zsl := new(zskiplist)
	//索引默认高度为1
	zsl.level = 1
	//跳表元素初始化为0
	zsl.length = 0
	//初始化一个头节点socre为0，元素为空
	zsl.header = zslCreateNode(ZSKIPLIST_MAXLEVEL, 0, nil)

	/**
	基于跳表最大高度32初始化头节点的索引，
	使得前驱指针指向null 跨度也设置为0
	*/
	for j = 0; j < ZSKIPLIST_MAXLEVEL; j++ {
		zsl.header.level[j].forward = nil
		zsl.header.level[j].span = 0
	}
	//头节点的前驱节点指向null，代表头节点之前没有任何元素
	zsl.header.backward = nil
	//初始化尾节点
	zsl.tail = nil
	return zsl
}

func zslCreateNode(level int, score float64, obj *robj) *zskiplistNode {
	zn := new(zskiplistNode)
	zn.level = make([]zskiplistLevel, level)
	zn.score = score
	zn.obj = obj
	return zn
}

func zslInsert(zsl *zskiplist, score float64, obj *robj) *zskiplistNode {
	//创建一个update数组，记录插入节点每层索引中小于该score的最大值
	update := make([]*zskiplistNode, ZSKIPLIST_MAXLEVEL)
	//记录各层索引走到小于score最大节点的跨区
	rank := make([]int64, ZSKIPLIST_MAXLEVEL)
	//x指向跳表走节点
	x := zsl.header
	var i int
	//从跳表当前最高层索引开始，查找每层小于当前score的节点的最大值节点
	for i = zsl.level - 1; i >= 0; i-- {
		//如果当前索引是最高层索引，那么rank从0开始算
		if i == zsl.level-1 {
			rank[i] = 0
		} else { //反之本层索引直接从上一层的跨度开始往后查找
			rank[i] = rank[i+1]
		}
		/**
		如果前驱节点不为空，且符合以下条件，则指针前移：
		1. 节点小于当前插入节点的score
		2. 节点score一致，且元素值小于或者等于当前score
		*/
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score || (x.level[i].forward.score == score && x.level[i].forward.obj.String() < obj.String())) {
			//记录本层索引前移跨度
			rank[i] += x.level[i].span
			//索引指针先前移动
			x = x.level[i].forward

		}
		//记录本层小于当前score的最大节点
		update[i] = x
	}
	//随机生成新插入节点的索引高度
	level := zslRandomLevel()
	/**
	如果大于当前索引高度，则进行初始化，将这些高层索引的update数组都指向header节点，跨度设置为跳表中的元素数
	意为这些高层索引小于插入节点的最大值就是header
	*/
	if level > zsl.level {
		for i := zsl.level; i < level; i++ {
			rank[i] = 0
			update[i] = zsl.header
			update[i].level[i].span = zsl.length
		}
		//更新一下跳表索引的高度
		zsl.level = level
	}
	//基于入参生成一个节点
	x = zslCreateNode(level, score, obj)
	//从底层到当前最高层索引处理节点关系
	for i = 0; i < level; i++ {
		//将小于当前节点的最大节点的forward指向插入节点x，同时x指向这个节点的前向节点
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x
		//维护x和update所指向节点之间的跨度信息
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = rank[0] - rank[i] + 1
	}
	/**
	考虑到当前插入节点生成的level小于当前跳表最高level的情况
	该逻辑会将这些区间的update索引中的元素到其前方节点的跨度＋1，即代表这些层级索引虽然没有指向x节点，
	但因为x节点插入的缘故跨度要加1
	*/
	for i = level; i < zsl.level; i++ {
		update[i].level[i].span++
	}
	//如果1级索引是header，则x后继节点不指向该节点，反之指向
	if update[0] == zsl.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}
	//如果x前向节点不为空，则让前向节点指向x
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else { //反之说明x是尾节点，tail指针指向它
		zsl.tail = x
	}
	//维护跳表长度信息
	zsl.length++
	return x
}

func zslRandomLevel() int {
	level := 1
	for rand.Float64() < ZSKIPLIST_P && level < ZSKIPLIST_MAXLEVEL {
		level++
	}
	return level
}

func zslGetRank(zsl *zskiplist, score float64, obj *robj) int64 {
	var rank int64
	//从索引最高节点开始进行查找
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		//如果前向节点不为空且score小于查找节点，或者score相等，但是元素字符序比值小于或者等于则前移，同时用rank记录跨度
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score || (x.level[i].forward.score == score && x.level[i].forward.obj.String() <= obj.String())) {
			rank += x.level[i].span
			x = x.level[i].forward
		}
		//上述循环结束，比对一直，则返回经过的跨度
		if x.obj != nil && x.obj.String() == obj.String() {
			return rank
		}
	}
	return 0
}

func zslDelete(zsl *zskiplist, score float64, obj *robj) int64 {
	update := make([]*zskiplistNode, ZSKIPLIST_MAXLEVEL)
	//找到每层索引要删除节点的前一个节点
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score || (x.level[i].forward.score == score && x.level[i].forward.obj.String() < obj.String())) {
			x = x.level[i].forward
		}
		update[i] = x
	}
	//查看1级索引前面是否就是要删除的节点，如果是则直接调用zslDeleteNode删除节点，并断掉前后节点关系
	x = x.level[0].forward
	if x != nil && x.obj.String() == obj.String() {
		zslDeleteNode(zsl, x, update)
		return 1
	}
	return 0
}

func zslDeleteNode(zsl *zskiplist, x *zskiplistNode, update []*zskiplistNode) {

	var i int
	for i = 0; i < zsl.level; i++ {
		/*
			如果索引前方就是删除节点，当前节点span为：
			当前节点到x +x到x前向节点 -1
		*/
		if update[i].level[i].forward == x {
			update[i].level[i].span += x.level[i].span - 1
			update[i].level[i].forward = x.level[i].forward
		} else {
			//反之说明该节点前方不是x的索引，直接减去x的跨步1即
			update[i].level[i].span -= 1
		}
	}
	//维护删除后的节点前后关系
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		zsl.tail = x.backward
	}
	//将全空层的索引删除
	for zsl.level > 1 && zsl.header.level[zsl.level-1].forward == nil {
		zsl.level--
	}
	//维护跳表节点信息
	zsl.length--

}

func zaddCommand(c *redisClient) {
	//传入0，即本次传入的score在元素存在情况下执行覆盖score而非累加score
	zaddGenericCommand(c, 0)
}

func zaddGenericCommand(c *redisClient, incr int) {
	//拿到有序集合的key
	key := c.argv[1]

	var ele *robj
	var zobj *robj
	var j uint64
	var score float64
	//初始化变量记录本次操作添加和更新的元素数
	var added int64
	var updated int64

	//参数非偶数，入参异常直接输出错误后返回
	if c.argc%2 != 0 {
		addReplyError(c, shared.syntaxerr)
		return
	}
	//减去zadd和key 再除去2 得到本次插入的元素数
	elements := (c.argc - 2) / 2

	//创建scores记录每个元素对应的score值
	scores := make([]float64, elements)
	for j = 0; j < elements; j++ {
		//对score进行转换，若报错直接返回
		if !getDoubleFromObjectOrReply(c, c.argv[2+j*2], &scores[j], nil) {
			return
		}
	}

	//若为空则创建一个有序集合,并添加到数据库中
	zobj = lookupKeyWrite(c.db, c.argv[1])
	if zobj == nil {
		zobj = createZsetObject()
		dbAdd(c.db, key, zobj)
	} else if zobj.robjType != REDIS_ZSET { //若类型不对则返回异常
		addReply(c, shared.wrongtypeerr)
		return
	}

	zs := (*zobj.ptr).(*zset)

	//基于元素数遍历集合
	for j = 0; j < elements; j++ {
		//拿到本次元素对应的score
		score = scores[j]
		//拿到对应的元素
		ele = c.argv[3+j*2]
		k := (*ele.ptr).(string)

		//如果该元素存在于字典中
		if zs.dict[k] != nil {
			//拿到当前元素对应的score
			curScore := zs.dict[k]
			//若不一样则更新字典中对应元素的score，并将该元素从跳表中删除再插入
			if *curScore != score {
				zslDelete(zs.zsl, *curScore, c.argv[3+j*2])
				zslInsert(zs.zsl, score, c.argv[3+j*2])
				zs.dict[k] = &score
				//维护更新数
				updated++
			}

		} else { //若是新增则插入到有序集合对应的跳表和字典中
			zslInsert(zs.zsl, score, c.argv[3+j*2])
			zs.dict[k] = &score
			//维护添加数
			added++
		}

	}

	//返回本次插入数
	addReplyLongLong(c, added)

}

func zcardCommand(c *redisClient) {
	//限定为有序集合是否存在且类型是否为有序集合
	zobj := lookupKeyReadOrReply(c, c.argv[1], shared.czero)
	if zobj == nil || checkType(c, zobj, REDIS_ZSET) {
		return
	}
	//拿到其底层的跳表返回元素数
	zs := (*zobj.ptr).(*zset)
	addReplyLongLong(c, zs.zsl.length)
}

func zrankCommand(c *redisClient) {
	zrankGenericCommand(c, 0)
}

func zrankGenericCommand(c *redisClient, reverse int) {
	//从参数中拿到有序集合的key和本次要查看排名的元素
	key := c.argv[1]
	ele := c.argv[2]

	//查看有序集合是否存在
	o := lookupKeyReadOrReply(c, key, nil)
	if o == nil || checkType(c, o, REDIS_ZSET) {
		return
	}
	//获取有序集合底层的跳表的长度
	zs := (*o.ptr).(*zset)
	llen := zs.zsl.length
	//查看元素在字典中是否存在
	k := (*ele.ptr).(string)
	score, exists := zs.dict[k]
	//如果存在则查看其在跳表中的排名
	if exists {
		//zslGetRank返回元素从头节点开始算经过的步数，例如aa是第一个元素，那么header走到它需要跨1步，所以返回1
		rank := zslGetRank(zs.zsl, *score, ele)
		//如果要返回倒叙结果则基于长度减去rank
		if reverse == 1 {
			addReplyLongLong(c, llen-rank)
		} else {
			//将rank减去1得到元素实际的索引值
			addReplyLongLong(c, rank-1)
		}
	} else { //不存在返回空
		addReply(c, shared.nullbulk)
	}

}

func zremCommand(c *redisClient) {
	var deleted int64
	//检查有序集合是否存在且类型是否是有序集合类型，如果为空或者类型不一致则返回
	o := lookupKeyWriteOrReply(c, c.argv[1], shared.czero)
	if o == nil || checkType(c, o, REDIS_ZSET) {
		return
	}
	zs := (*o.ptr).(*zset)

	var j uint64
	//遍历元素
	for j = 2; j < c.argc; j++ {
		//拿到元素字符串
		ele := (*c.argv[j].ptr).(string)
		//如果不为空则将其从底层字典和跳表中删除
		if zs.dict[ele] != nil {
			//更新删除结果
			deleted++
			zslDelete(zs.zsl, *zs.dict[ele], c.argv[j])
			delete(zs.dict, ele)

			//如果发现字典为空，说明有序集合没有元素了，直接将该有序集合从字典中期删除
			if len(zs.dict) == 0 {
				dbDelete(c.db, c.argv[1])
			}
		}
	}
	//返回删除数
	addReplyLongLong(c, deleted)

}
