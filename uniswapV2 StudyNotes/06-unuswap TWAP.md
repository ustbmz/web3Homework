老师板子坏掉的那段，就是从「时间加权平均」推到「用累加器差商」那一步。补完整如下。

## 1. 价格是阶梯函数

Uniswap V2 的现货价只在交易发生时变。从 \(t_i\) 到 \(t_{i+1}\) 之间，储备不变，价格就钉死在：

$$
P_i = \frac{\text{reserve}_1}{\text{reserve}_0}
$$

时间长度记：

$$
\Delta t_i = t_{i+1} - t_i
$$

图上那条阶梯线，每一格就是一个 \(P_i\)，宽度是 \(\Delta t_i\)。

## 2. 从 \(t_0\) 到 \(t_n\) 的平均价

「按时间加权」就是：哪个价格钉得越久，权重越大。

$$
\text{TWAP}(t_0,t_n)
= \frac{(t_1-t_0)P_0 + (t_2-t_1)P_1 + \cdots + (t_n-t_{n-1})P_{n-1}}{t_n-t_0}
$$

把分母拆进每一项：

$$
= \frac{\Delta t_0}{t_n-t_0}P_0 + \frac{\Delta t_1}{t_n-t_0}P_1 + \cdots + \frac{\Delta t_{n-1}}{t_n-t_0}P_{n-1}
$$

写成求和：

$$
\text{TWAP}(t_0,t_n) = \sum_{i=0}^{n-1} \frac{\Delta t_i}{t_n-t_0}\, P_i
$$

这就是图左下角那张。

## 3. 任意区间 \([t_k, t_n]\)

老师板子断在这里。把起点从 \(t_0\) 换成任意的 \(t_k\)，只保留 \(t_k\) 到 \(t_n\) 之间那些台阶：

$$
\text{TWAP}(t_k,t_n)
= \frac{\Delta t_k\, P_k + \Delta t_{k+1}\, P_{k+1} + \cdots + \Delta t_{n-1}\, P_{n-1}}{t_n-t_k}
$$

$$
= \sum_{i=k}^{n-1} \frac{\Delta t_i}{t_n-t_k}\, P_i
$$

这就是图右下角圈起来的式子。

## 4. 关键累加器（板子坏掉后最关键的一步）

把求和式分子单独拿出来：

$$
\sum_{i=k}^{n-1} \Delta t_i\, P_i
$$

定义一个从 \(t_0\) 一直往前累加的量：

$$
C(t_m) = \sum_{i=0}^{m-1} \Delta t_i\, P_i
$$

它就是 Uniswap 里的 `priceCumulative`。那么任意区间分子正好是两端累加器的差：

$$
\sum_{i=k}^{n-1} \Delta t_i\, P_i = C(t_n) - C(t_k)
$$

代回 TWAP：

$$
\text{TWAP}(t_k,t_n) = \frac{C(t_n) - C(t_k)}{t_n - t_k}
$$

这就是顶部幻灯片的公式：

$$
\text{TWAP} = \frac{\text{priceCumulative}_2 - \text{priceCumulative}_1}{\text{timestamp}_2 - \text{timestamp}_1}
$$

数字验算：

$$
\frac{48120 - 11400}{1583535828 - 1583532228} = \frac{36720}{3600} = 10.2
$$

## 5. 合约里怎么维护 \(C(t)\)

每次 `mint` / `burn` / `swap` / `sync` 都会进 `_update`。如果距离上次更新过了 `timeElapsed` 秒，就把「旧价格 × 这段时间」加进累加器，再写入新储备：

```77:80:contracts/UniswapV2Pair.sol
        if (timeElapsed > 0 && _reserve0 != 0 && _reserve1 != 0) {
            price0CumulativeLast += uint(UQ112x112.encode(_reserve1).uqdiv(_reserve0)) * timeElapsed;
            price1CumulativeLast += uint(UQ112x112.encode(_reserve0).uqdiv(_reserve1)) * timeElapsed;
        }
```

也就是：

$$
C_{\text{new}} = C_{\text{old}} + P_{\text{old}} \cdot \Delta t
$$

价格只在交易时变，中间空窗期用旧价继续计时，正好对应阶梯函数。

读 TWAP 时不需要每次都 `sync`。Oracle 库会把「从上次更新到现在」这段也按当前储备虚拟累加一遍，得到此刻的 \(C(t)\)，再和你之前存下的 \(C(t_k)\) 做差商。

## 一句话串起来

时间加权平均 = 各段价格 × 持续时间的和，再除以总时长。  
Uniswap 把「价格 × 持续时间」一路累加成 `priceCumulative`，所以任意区间的 TWAP 只要两端各拍一次快照，做一次减法除法就行，不用重放中间每一笔交易。