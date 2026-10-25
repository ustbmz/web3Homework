# Solidity 智能合约开发学习大纲

依据桌面课程目录 `/Users/cnmz/Desktop/solidity` 中的 18 份课件整理。课程名是「Solidity 智能合约开发：从入门到精通」，讲师 Layer。建议按下面的课次顺序学，不要跳着看安全课和 Gas 课。前面的存储位置、类型和函数修饰符，后面每一课都在用。

学习时每课做三件事：先看这一课「要会什么」，再对着课件把示例在 Remix 里跑一遍，最后用「注意」里的点检查自己的代码。

课件只覆盖「把合约写出来」。文末「Web3 还需要学什么」是我按这条路线补的后续建议，不在这 18 份 PDF 里。

---

## 课程地图


| 顺序  | 课件          | 主题          | 学完要能做什么                                  |
| --- | ----------- | ----------- | ---------------------------------------- |
| 1   | lesson1.1   | 合约基础与环境搭建   | 在 Remix 里写出、编译、部署一个最小合约                  |
| 2   | lesson1.2   | EVM 存储结构    | 正确选择 storage、memory、calldata             |
| 3   | lesson2.1   | 数据类型基础      | 选对整数、地址、字节和字符串                           |
| 4   | lesson3.1   | 数组          | 增删改查数组，并知道删除时的坑                          |
| 5   | lesson3.2   | 映射和结构体      | 用 mapping + struct + 数组存可查询、可遍历的数据       |
| 6   | lesson3.3   | 函数与修饰符      | 写对可见性和 view / pure / payable，会写 modifier |
| 7   | lesson4.1   | 控制流语句       | 写条件、循环和 require，并控制循环的 Gas               |
| 8   | lesson4.2   | 特殊类型与全局变量   | 安全转账，分清 msg.sender 和 tx.origin           |
| 9   | lesson5.1   | 简单代币合约      | 实现 ERC20 的转账和授权                          |
| 10  | lesson6.1   | 合约继承        | 用继承、virtual / override、抽象合约和接口复用代码       |
| 11  | lesson6.2   | 库合约 Library | 把无状态工具函数放进 library                       |
| 12  | lesson7.1   | 事件 Events   | 用事件把链上动作通知给链下                            |
| 13  | lesson7.2   | 错误处理与自定义错误  | 分清 require、assert、revert 和自定义 error      |
| 14  | lesson8.1   | 合约间调用       | 用接口和 call / delegatecall 调用别的合约          |
| 15  | lesson8.2   | 合约安全基础      | 防住重入、越权、拒绝服务和抢跑                          |
| 16  | lesson9.1   | Gas 优化      | 从存储、类型和函数三方面省 Gas                        |
| 17  | lesson9.2   | 设计模式        | 会选用访问控制、提现、状态机、代理、工厂、紧急停止                |
| 18  | lessonA10.1 | NFT 市场      | 把 ERC721、上架、购买、版税和拍卖串成一个市场               |


---



## 1. 合约基础与环境搭建

**要会什么**

- 智能合约是部署在区块链上的程序。部署之后代码改不了，执行过程公开，条件满足就自动执行。
- 合约跑在以太坊虚拟机（EVM）上。Solidity 源码先编译成字节码，再部署到链上，由全球节点执行。
- Gas 是计算量的计量单位。循环、写存储都很贵，无限循环会被 Gas 耗尽而回滚。
- 用 [Remix](https://remix.ethereum.org) 完成编译、部署和调用。学习阶段选 Remix VM，不花真钱。

一个合约的骨架是：

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract HelloWorld {
    string public message;

    constructor() {
        message = "Hello, Solidity!";
    }

    function updateMessage(string memory newMessage) public {
        message = newMessage;
    }
}
```

**注意**

- `pragma solidity ^0.8.0` 里的 `^` 表示 `0.8.0` 到 `0.9.0`（不含）。0.8 会检查整数溢出，不要随便降到 0.7。
- `constructor` 只在部署时执行一次。记录创建者要在这里写 `owner = msg.sender`。
- `public` 状态变量会自动生成同名读取函数。
- 本课作业是给 HelloWorld 加上 `address public owner` 和 `getOwner()`。

**示例：带所有权的消息合约**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract OwnedMessage {
    address public immutable owner;
    string public message;

    constructor(string memory initialMessage) {
        owner = msg.sender;
        message = initialMessage;
    }

    function updateMessage(string calldata newMessage) external {
        require(msg.sender == owner, "not owner");
        message = newMessage;
    }

    function getOwner() external view returns (address) {
        return owner;
    }
}
```

部署时在参数框中输入初始消息，只有部署账户可以调用 `updateMessage`。

---



## 2. EVM 存储结构

**要会什么**

EVM 里和写代码直接相关的是三种数据位置：


| 位置       | 生命周期     | 能不能改 | 什么时候用                |
| -------- | -------- | ---- | -------------------- |
| storage  | 永久写在链上   | 能    | 状态变量，必须跨交易保存的数据      |
| memory   | 这次函数调用期间 | 能    | 函数里的临时数组、字符串、需要修改的参数 |
| calldata | 这次调用期间   | 不能   | 外部函数里只读的数组和字符串参数     |


另外还有栈（stack），局部的值类型大多放在栈上，编译器自己管。

**注意**

- 新开辟一个 storage 槽大约 20,000 Gas，改已有非零值大约 5,000 Gas，读一次冷存储大约 800 Gas。memory 大约每字节约 3 Gas。
- 外部函数如果只读参数，用 `calldata`，不要先复制到 `memory`。
- 循环里不要反复读 `array.length` 或同一个 storage 变量，先抄到局部变量。
- 小类型要挨着声明才能打包进同一个 32 字节槽。`uint8` 和 `uint256` 交错声明会各自占一个槽。mapping 和动态数组不能参与打包。
- `constant` 和 `immutable` 不占 storage 槽。编译期就能确定的用 `constant`，部署时才能确定的（比如 owner）用 `immutable`。
- 选择顺序：要永久保存就用 storage；外部只读参数用 calldata；要在函数里改的临时数据用 memory。

**示例：修改引用与修改副本**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract DataLocationsDemo {
    uint256[] public numbers;

    constructor() {
        numbers.push(10);
        numbers.push(20);
    }

    function changeStored(uint256 value) external {
        uint256[] storage ref = numbers;
        ref[0] = value;
    }

    function changeCopy(uint256 value) external view returns (uint256[] memory) {
        uint256[] memory copy = numbers;
        copy[0] = value;
        return copy;
    }

    function sum(uint256[] calldata values) external pure returns (uint256 total) {
        require(values.length <= 100, "too many values");
        for (uint256 i = 0; i < values.length; i++) {
            total += values[i];
        }
    }
}
```

在 Remix 中分别验证：

- 部署后调用 `numbers(0)`，返回 `10`。
- 调用 `changeCopy(99)`，返回 `[99, 20]`，但 `numbers(0)` 仍为 `10`：storage 到 memory 的赋值会复制数据。
- 调用 `changeStored(99)` 后，`numbers(0)` 变成 `99`：局部 storage 引用会修改原始状态。
- 调用 `sum([1,2,3])`，返回 `6`，参数只读，无须复制到 memory。

---



## 3. 数据类型基础

**要会什么**

- 值类型赋值时复制一份，改副本不影响原值：`bool`、`int` / `uint`、`address`、`bytes1` 到 `bytes32`、`enum`。
- 引用类型传递的是同一份数据：数组、`string`、`bytes`、`struct`、`mapping`。在 memory 里 `arr2 = arr1` 之后，改 `arr2[0]` 会同时改到 `arr1[0]`。
- 整数以 8 位为步长，从 8 到 256。日常计算优先 `uint256`，它和 EVM 字长一致。需要把多个小字段塞进一个槽时，才改用 `uint8` 这类小类型。
- `address` 不能收款。要收款或调用 `transfer`，必须是 `address payable`，转换写法是 `payable(addr)`。
- `bytes1` 到 `bytes32` 是定长值类型，适合哈希和函数选择器。`bytes` 和 `string` 是变长引用类型。

**注意**

- Solidity 0.8 及之后，加减乘溢出默认回滚。0.8 之前会静默绕回，那是真实漏洞。
- `string` 不能用 `==` 比内容。要比内容，先 `keccak256(abi.encodePacked(s))` 再比较哈希。
- 大类型转小类型会截断。先 `require(value <= type(uint8).max)`，再 `uint8(value)`。
- `&&` 和 `||` 会短路。便宜、容易失败的检查放前面，贵的外部调用放后面。
- 只有有限个命名状态时用 `enum`，不要用 `0`、`1`、`2` 这种魔术数字。

**示例：值类型、字符串与安全转换**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract TypesDemo {
    enum Status { Pending, Finished }

    function copyValue(uint256 original) external pure returns (uint256, uint256) {
        uint256 copy = original;
        copy += 1;
        return (original, copy);
    }

    function sameText(string calldata a, string calldata b) external pure returns (bool) {
        return keccak256(bytes(a)) == keccak256(bytes(b));
    }

    function toUint8(uint256 value) external pure returns (uint8) {
        require(value <= type(uint8).max, "out of range");
        return uint8(value);
    }

    function chooseStatus(bool finished) external pure returns (Status) {
        return finished ? Status.Finished : Status.Pending;
    }
}
```

调用 `copyValue(5)` 返回 `(5, 6)`；`sameText("abc", "abc")` 返回 `true`；`toUint8(256)` 会回滚，`toUint8(255)` 返回 `255`。

---



## 4. 数组

**要会什么**

- 定长数组 `uint[5]` 长度不能变，不能 `push` / `pop`。
- 动态数组 `uint[]` 存在 storage 里时可以 `push` 和 `pop`。
- memory 数组必须 `new uint[](长度)`，长度创建后不能变，也不能 `push`。
- `arr[i]` 读写单个元素，`arr.length` 取长度。

**注意**

- `delete arr[i]` 只是把该位置成 0，**长度不变**，后面的元素也不会前移。数组里会留下空洞。
- 要真正删掉一个元素，有两种做法。要保持顺序：从该下标起后面的元素依次前移，再 `pop`。不要求顺序：用最后一个元素覆盖这个下标，再 `pop`。第二种更省 Gas。
- 声明二维数组时，顺序和常见语言相反。`uint[3][4]` 是 4 个长度为 3 的数组。
- 链上不要遍历无上限的大数组。单次交易的 Gas 有上限，数组太长交易会失败。给数组设最大长度，或者改成按页处理。
- 需要按下标顺序访问、数量不大时用数组。需要按地址或 id 立刻查到时用 mapping。

**示例：数组增删改查与两种删除方式**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract ArraysDemo {
    uint256[] public numbers;
    uint256 public constant MAX_LENGTH = 100;

    function add(uint256 value) external {
        require(numbers.length < MAX_LENGTH, "array full");
        numbers.push(value);
    }

    function update(uint256 index, uint256 value) external {
        require(index < numbers.length, "invalid index");
        numbers[index] = value;
    }

    function clear(uint256 index) external {
        require(index < numbers.length, "invalid index");
        delete numbers[index];
    }

    function removeUnordered(uint256 index) external {
        require(index < numbers.length, "invalid index");
        numbers[index] = numbers[numbers.length - 1];
        numbers.pop();
    }

    function removeOrdered(uint256 index) external {
        uint256 length = numbers.length;
        require(index < length, "invalid index");
        for (uint256 i = index; i < length - 1; i++) {
            numbers[i] = numbers[i + 1];
        }
        numbers.pop();
    }

    function getAll() external view returns (uint256[] memory) {
        return numbers;
    }
}
```

先依次 `add(10)`、`add(20)`、`add(30)`，再调用 `getAll()`：

- `clear(1)` 得到 `[10, 0, 30]`，长度仍是 3。
- 在重新部署并加入相同数据后，`removeUnordered(0)` 得到 `[30, 20]`。
- 在重新部署并加入相同数据后，`removeOrdered(0)` 得到 `[20, 30]`。

示例允许任何账户修改数组，仅用于学习；实际项目需根据业务限制写入权限。限制最大长度也使保序删除的循环有界。

---



## 5. 映射和结构体

**要会什么**

- `mapping(keyType => valueType)` 是键值对，查找是 O(1)。存储位置由 `keccak256(key, 槽位)` 算出来。
- 键没有单独存，所以 mapping 不能遍历，没有 `length`，不能整体 `delete`，也不能当函数参数或返回值。它只能作为 storage 状态变量。
- 没写过的键也会“存在”，读出来是该类型的默认值：数字是 0，布尔是 false，地址是零地址。
- 嵌套 mapping `mapping(address => mapping(address => uint256))` 就是 ERC20 的授权表：`allowance[所有者][被授权人]`。
- `struct` 把多个字段收成一条记录。可以逐字段赋值，也可以用键值方式一次性构造。
- 标准组合是：`mapping` 负责按键查找，`struct` 负责打包字段，再加一个数组保存所有键，这样才遍历得了。

**注意**

- 余额为 0 可能是“真的是 0”，也可能是“这个人从没登记过”。需要区分时，在 struct 里加 `bool exists`。
- 重置所有人的票数、列出所有捐赠者，都不能只靠 mapping。必须自己维护一份地址或名字数组。这就是投票合约里 `candidates` 数组存在的原因。
- `delete map[key]` 只清这一个键，把它变回默认值。

**示例：mapping + struct + 地址列表**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract MembersDemo {
    struct Member {
        string name;
        uint256 points;
        bool exists;
    }

    mapping(address => Member) public members;
    address[] public memberAddresses;

    function register(string calldata name) external {
        require(!members[msg.sender].exists, "already registered");
        require(bytes(name).length > 0 && bytes(name).length <= 64, "invalid name");
        require(memberAddresses.length < 100, "member limit");
        members[msg.sender] = Member({name: name, points: 0, exists: true});
        memberAddresses.push(msg.sender);
    }

    function updateName(string calldata name) external {
        Member storage member = members[msg.sender];
        require(member.exists, "not registered");
        require(bytes(name).length > 0 && bytes(name).length <= 64, "invalid name");
        member.name = name;
    }

    function memberCount() external view returns (uint256) {
        return memberAddresses.length;
    }
}
```

调用 `register("Alice")` 后，用当前账户地址查询 `members`，得到 `("Alice", 0, true)`。重复登记会回滚，`memberCount()` 仍是 `1`；换一个账户登记后变成 `2`。未登记地址返回 `("", 0, false)`，可用 `exists` 区分零积分会员和未登记用户。

---



## 6. 函数与修饰符

**要会什么**

函数的形状是：

```solidity
function 名字(参数) 可见性 状态修饰符 自定义修饰符 returns (返回类型) {
    // 函数体
}
```

可见性：


| 修饰符      | 外部能调 | 内部能调                | 子合约能调 |
| -------- | ---- | ------------------- | ----- |
| public   | 能    | 能                   | 能     |
| external | 能    | 不能直接调，要写 `this.f()` | 能     |
| internal | 不能   | 能                   | 能     |
| private  | 不能   | 能                   | 不能    |


状态修饰符：


| 修饰符     | 读状态 | 改状态 | 收以太币 |
| ------- | --- | --- | ---- |
| 默认（不写）  | 能   | 能   | 不能   |
| view    | 能   | 不能  | 不能   |
| pure    | 不能  | 不能  | 不能   |
| payable | 能   | 能   | 能    |


自定义 `modifier` 在函数体之前插入检查，用 `_` 表示“接着执行函数体”。多个 modifier 按书写顺序执行。

同名函数可以重载，参数类型或个数必须不同。

**注意**

- 每个函数都必须写可见性，漏了会编译失败。
- 对外、参数又是大数组或字符串时，优先 `external` 配 `calldata`，比 `public` 省一次内存复制。
- `pure` 连 `msg.sender`、`msg.value`、`block.timestamp` 都不能读。只处理参数的工具函数才是 `pure`。
- 没收 `payable` 的函数，调用时不能附带以太币。
- `onlyOwner` 这类检查放在 modifier 里，不要在每个函数里复制一遍 `require(msg.sender == owner)`。
- 可见性是安全边界。不该给用户调的函数不要写成 `public`。

**示例：可见性、view / pure 与 onlyOwner**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract FunctionsDemo {
    address public immutable owner;
    uint256 private count;

    constructor() {
        owner = msg.sender;
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "not owner");
        _;
    }

      function increment(uint256 amount) external onlyOwner {
        require(amount > 0, "zero amount");
        _increase(amount);
    }

    function _increase(uint256 amount) internal {
        count += amount;
    }

    function getCount() public view returns (uint256) {
        return count;
    }

    function add(uint256 a, uint256 b) external pure returns (uint256) {
        return a + b;
    }
}
```

部署账户调用 `increment(3)` 后，`getCount()` 返回 `3`；换账户调用 `increment` 会回滚。`add(2, 5)` 返回 `7`，只使用参数，因此可以声明为 `pure`。`_increase` 不作为外部接口暴露，但子合约可调用它。

这里不接收 ETH，避免示例资金没有提款路径；`payable` 的收款与提款练习可结合第 8 课和讨饭合约完成。`private` 限制代码访问，不会让链上数据保密。

---



## 7. 控制流语句

**要会什么**

- 顺序、`if` / `else if` / `else`、三元运算符 `条件 ? a : b`。
- `for` 和 `while`。`else if` 按书写顺序匹配，第一个成立的分支执行后，后面的不再看。
- 错误处理三件套，下一课的自定义错误会再加深：
  - `require`：检查外部输入和调用前提，失败是用户或调用方的问题。
  - `assert`：检查内部不变量，失败说明合约自己有 bug。
  - `revert`：在复杂分支里主动回滚。

**注意**

- 循环次数必须有上限。无界循环会把 Gas 烧光，交易回滚，功能等于不可用。
- 能用 mapping 一次查到的，不要写循环去找。
- 条件复杂时先用提前 `return` 或 modifier 挡住非法情况，避免一层套一层。
- 投票、众筹这类流程，用 `enum` 表示当前阶段，每个函数开头先检查阶段对不对。

**示例：分支、有界循环与状态机**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract ControlFlowDemo {
    enum Stage { Open, Closed }
    Stage public stage;
    address public immutable owner;
    uint256 public submissions;

    constructor() {
        owner = msg.sender;
    }

    function grade(uint256 score) external pure returns (string memory) {
        require(score <= 100, "invalid score");
        if (score >= 90) {
            return "A";
        } else if (score >= 60) {
            return "Pass";
        } else {
            return "Fail";
        }
    }

    function max(uint256 a, uint256 b) external pure returns (uint256) {
        return a > b ? a : b;
    }

    function sum(uint256[] calldata values) external pure returns (uint256 total) {
        require(values.length <= 100, "too many values");
        for (uint256 i = 0; i < values.length; i++) {
            total += values[i];
        }
    }

    function halveUntilZero(uint256 value) external pure returns (uint256 steps) {
        while (value > 0) {
            value /= 2;
            steps++;
        }
        assert(steps <= 256);
    }

    function submit() external {
        require(stage == Stage.Open, "closed");
        submissions++;
    }

    function close() external {
        require(msg.sender == owner, "not owner");
        if (stage == Stage.Closed) {
            revert("already closed");
        }
        stage = Stage.Closed;
    }
}
```

验证方式：

- `grade(95)` 返回 `"A"`，`grade(60)` 返回 `"Pass"`，`grade(101)` 回滚。
- `max(3, 8)` 返回 `8`；`sum([1,2,3])` 返回 `6`。
- `halveUntilZero(8)` 返回 `4`。每次整数除以 2，uint256 最多经过 256 次就会变成 0，所以这个 while 有界；`assert` 验证这个内部不变量。
- 初始 `stage` 为 `0`（Open），调用 `submit()` 可以增加计数。部署账户调用 `close()` 后，stage 为 `1`（Closed），再次 submit 或 close 会回滚。

第 1 至第 7 课的新示例均可独立复制到 `.sol` 文件，用 Remix 的 Solidity 0.8.20 或更高的 0.8.x 编译器和 Remix VM 练习。示例侧重语法演示，生产项目还需补充业务约束和测试。

---



## 8. 特殊类型与全局变量

**要会什么**

全局对象：

- `msg.sender`：这一次调用的直接调用者。权限判断用它。
- `msg.value`：这次调用附带的 wei。只有 `payable` 函数里才有意义。
- `address(this).balance`：本合约当前持有的以太币。
- `block.timestamp`：当前区块时间，适合截止时间，不适合精确到秒的随机数。
- `tx.origin`：整笔交易最初的外部账户。不要用它做权限判断。

三种转账：


| 方法                        | Gas           | 失败时               |
| ------------------------- | ------------- | ----------------- |
| `transfer`                | 固定 2300       | 自动回滚              |
| `send`                    | 固定 2300       | 返回 `false`，自己必须检查 |
| `call{value: amount}("")` | 可以把剩余 Gas 转过去 | 返回 `bool`，必须检查    |


**注意**

- 普通 `address` 没有 `transfer`。先转成 `payable`。
- `send` 失败不会自动回滚。不检查返回值，账已经改了，钱却没转出去。
- 课件后半段推荐 `call` 并检查返回值，因为接收方如果是合约，2300 Gas 常常不够执行它的接收逻辑。作业里如果明确要求 `transfer`，按作业写。
- 先改自己的账，再给外部转钱。这就是 CEI：Checks 检查，Effects 改状态，Interactions 最后才外部调用。顺序反了就会被重入。
- `msg.sender` 是直接调用者，合约调合约时会变。`tx.origin` 一直是最外层的人。用 `tx.origin == owner` 做权限，中间夹一个恶意合约就能骗过检查。

**示例：收款、查询调用信息与自行提款**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract EtherWalletDemo {
    mapping(address => uint256) public deposits;

    function deposit() external payable {
        require(msg.value > 0, "zero deposit");
        deposits[msg.sender] += msg.value;
    }

    function context() external view returns (address, address, uint256, uint256) {
        return (msg.sender, tx.origin, block.timestamp, address(this).balance);
    }

    function withdraw(uint256 amount) external {
        require(amount > 0 && amount <= deposits[msg.sender], "invalid amount");
        deposits[msg.sender] -= amount;
        (bool success, ) = payable(msg.sender).call{value: amount}("");
        require(success, "transfer failed");
    }
}
```

在 Remix VM 中选择账户 A，将 Value 设为 `1 ether` 后调用 `deposit()`，再把 Value 重置为 `0`。查询 `deposits(A)` 和 `context()` 中的合约余额，都应为 `1000000000000000000` wei。调用 `withdraw(400000000000000000)` 后，A 的记录和合约余额都剩 `600000000000000000` wei。

换账户 B，不能提取 A 的存款。直接由钱包调用时，`msg.sender` 和 `tx.origin` 相同；经其他合约调用时可能不同。这里仅展示 `tx.origin`，不用它鉴权。提款先扣账再外部调用，转账失败时 `require` 会让扣账一起回滚。

---



## 9. 简单代币合约（ERC20）

**要会什么**

ERC20 是同质化代币标准。核心是 6 个函数和 2 个事件：

- 查询：`totalSupply()`、`balanceOf(account)`、`allowance(owner, spender)`
- 转账：`transfer(to, amount)` 从调用者自己的余额转出
- 授权：`approve(spender, amount)` 允许别人花自己的币，新额度覆盖旧额度
- 代花：`transferFrom(from, to, amount)` 花授权额度，同时扣减 `allowance`
- 事件：`Transfer`、`Approval`

扩展功能是 `mint`（增发）和 `burn`（销毁），这两项必须限制调用者，不能让任何人铸造。

**注意**

- `transfer` 要拒绝零地址，并检查余额够不够。
- `transferFrom` 要同时检查余额和授权额度，两本账一起改。
- `approve` 是覆盖，不是累加。前端如果先查再改，中间可能被抢跑改掉旧额度。
- 余额用 `mapping(address => uint256)`。总供应量单独用一个 `uint256` 状态变量。

**示例：最小 ERC20 余额与授权账本**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract TokenDemo {
    string public constant name = "Learning Token";
    string public constant symbol = "LRN";
    uint8 public constant decimals = 18;
    uint256 public totalSupply;
    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    constructor() {
        totalSupply = 1000 * 10 ** uint256(decimals);
        balanceOf[msg.sender] = totalSupply;
        emit Transfer(address(0), msg.sender, totalSupply);
    }

    function transfer(address to, uint256 amount) external returns (bool) {
        _transfer(msg.sender, to, amount);
        return true;
    }

    function approve(address spender, uint256 amount) external returns (bool) {
        require(spender != address(0), "zero spender");
        allowance[msg.sender][spender] = amount;
        emit Approval(msg.sender, spender, amount);
        return true;
    }

    function transferFrom(address from, address to, uint256 amount) external returns (bool) {
        uint256 allowed = allowance[from][msg.sender];
        require(allowed >= amount, "insufficient allowance");
        allowance[from][msg.sender] = allowed - amount;
        _transfer(from, to, amount);
        return true;
    }

    function _transfer(address from, address to, uint256 amount) internal {
        require(to != address(0), "zero recipient");
        require(balanceOf[from] >= amount, "insufficient balance");
        balanceOf[from] -= amount;
        balanceOf[to] += amount;
        emit Transfer(from, to, amount);
    }
}
```

部署账户 A 获得 1000 个代币，接口中的金额均为最小单位，`1 LRN = 10^18` 个最小单位：

1. A 调用 `transfer(B, 1000000000000000000)`，B 得到 1 LRN。
2. A 调用 `approve(B, 2000000000000000000)`，授权 B 花费 2 LRN。
3. 切换到 B，调用 `transferFrom(A, C, 1000000000000000000)`，C 得到 1 LRN，`allowance(A, B)` 剩 1 LRN。
4. 超出余额或授权额度的调用会回滚，包括此前对 allowance 的修改。

这里自动生成的 getter 提供 `totalSupply()`、`balanceOf(address)` 和 `allowance(address,address)`。示例固定初始供应量，未提供 mint/burn；所有授权包括最大整数都按实际花费扣减。生产代币优先使用 OpenZeppelin ERC20，并另行设计发行权限。

---



## 10. 合约继承

**要会什么**

- `contract Child is Parent` 是单继承，`is A, B` 是多继承。多继承的函数查找顺序用 C3 线性化，继承列表的顺序有意义。
- 父合约构造函数先执行，多个父合约按继承顺序从左到右。
- `virtual` 表示允许子合约重写，`override` 表示这里在重写父合约函数。两个都要写对，否则编译失败。
- `super.foo()` 按线性化顺序调用上一层，不是简单的“直接父合约”。
- `abstract contract` 可以有部分实现，至少有一个未实现函数，不能单独部署。
- `interface` 只有函数声明，没有实现，也没有状态变量。调别人的合约时，先写接口。

**注意**

- 能复用就继承，不要把同一段 `onlyOwner` 复制到每个合约。实际项目优先用 OpenZeppelin 的 `Ownable`、`ERC20`。
- 重写函数时，可见性不能比父函数更严。
- 子合约如果改了父合约的状态假设，要连父合约的检查一起看，不要只看子合约这一个函数。

**示例：抽象父合约、重写与接口调用**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

interface ICounterDemo {
    function getCount() external view returns (uint256);
}

abstract contract CounterBase {
    address public immutable owner;
    uint256 internal count;

    constructor() {
        owner = msg.sender;
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "not owner");
        _;
    }

    function increment() public virtual onlyOwner {
        count += 1;
    }

    function label() external pure virtual returns (string memory);
}

contract DoubleCounterDemo is CounterBase, ICounterDemo {
    function increment() public override onlyOwner {
        super.increment();
        count += 1;
    }

    function getCount() external view override returns (uint256) {
        return count;
    }

    function label() external pure override returns (string memory) {
        return "double counter";
    }
}

contract CounterReaderDemo {
    function read(address target) external view returns (uint256) {
        return ICounterDemo(target).getCount();
    }
}
```

在 Remix 选择并部署 `DoubleCounterDemo`，部署账户调用一次 `increment()` 后，`getCount()` 返回 `2`：`super.increment()` 增加 1，子合约再增加 1。换账户调用会回滚。父合约保留的权限检查也会执行。

再部署 `CounterReaderDemo`，将计数器地址传入 `read`，也返回 `2`。接口描述目标合约的调用方式，不验证目标是否可信。`CounterBase` 有未实现的 `label()`，因此不能单独部署。

---



## 11. 库合约 Library

**要会什么**

- `library` 用来放可复用、无状态的函数。库里不能声明状态变量。
- `using Lib for uint256` 之后，可以写成 `x.add(y)` 这种成员调用。
- `internal` 库函数会嵌进调用方的字节码，跳转方式和内部函数一样。
- `public` 库函数要单独部署，通过 `DELEGATECALL` 执行，这时用的是调用方的 storage。

**注意**

- 库不是用来存余额、owner 的。那些必须放在使用库的合约里。
- 0.8 已经自带溢出检查，新代码不必再为了安全去用 SafeMath。旧合约或刻意省 Gas 时才会看到它。
- 库函数尽量是 `pure` 或 `view`。需要改调用方状态时，把 storage 引用当参数传进去，并确认自己知道改的是谁的存储。

**示例：using for 与无状态库函数**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

library NumberTools {
    function min(uint256 a, uint256 b) internal pure returns (uint256) {
        return a < b ? a : b;
    }

    function clamp(uint256 value, uint256 lower, uint256 upper) internal pure returns (uint256) {
        require(lower <= upper, "invalid bounds");
        if (value < lower) return lower;
        if (value > upper) return upper;
        return value;
    }
}

contract LibraryDemo {
    using NumberTools for uint256;

    function smaller(uint256 a, uint256 b) external pure returns (uint256) {
        return a.min(b);
    }

    function inRange(uint256 value, uint256 lower, uint256 upper) external pure returns (uint256) {
        return value.clamp(lower, upper);
    }
}
```

部署 `LibraryDemo`，调用 `smaller(8, 3)` 返回 `3`；`inRange(15, 0, 10)` 返回 `10`；`inRange(5, 0, 10)` 返回 `5`；上下界颠倒会回滚。

`a.min(b)` 等价于 `NumberTools.min(a, b)`，左侧变量成为第一个参数。本例库函数均为 internal，无须单独部署库，也不涉及外部库链接；两个定义放在同一个文件即可。

---



## 12. 事件 Events

**要会什么**

```solidity
event Transfer(address indexed from, address indexed to, uint256 value);
emit Transfer(msg.sender, to, amount);
```

- 事件写进交易日志，给前端、索引服务和区块浏览器看。合约自己不能再把历史事件读回来。
- `indexed` 参数进 `topics`，可以被过滤。普通事件最多 3 个 `indexed`。
- `anonymous` 事件不把事件签名放进 topic，因此最多 4 个 `indexed`，也更省 Gas，但链下过滤更麻烦。

**注意**

- 余额、所有权这种合约逻辑还要读的数据，必须放 storage。事件只是通知，不能代替账本。
- 转账、授权、投票、上下架都要打事件，否则前端无法可靠地同步。
- 参数不要全部 `indexed`。需要过滤的地址、id 才标 `indexed`，金额放在普通数据里。

**示例：状态入库、历史发事件**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract MessageEventsDemo {
    mapping(address => string) public latestMessage;

    event MessageUpdated(address indexed author, uint256 indexed revision, string message);
    mapping(address => uint256) public revisions;

    function updateMessage(string calldata message) external {
        require(bytes(message).length > 0, "empty message");
        require(bytes(message).length <= 256, "message too long");
        latestMessage[msg.sender] = message;
        uint256 revision = ++revisions[msg.sender];
        emit MessageUpdated(msg.sender, revision, message);
    }
}
```

调用 `updateMessage("hello")` 后，`latestMessage(当前地址)` 返回 `"hello"`，`revisions(当前地址)` 返回 `1`。再调用 `updateMessage("world")`，当前状态变为 `"world"` 和 `2`；两次消息都能在 Remix 对应交易的 Logs 中查看。`author`、`revision` 可供链下按 topic 过滤，`message` 作为普通日志数据存储；合约只能通过 storage 读取最新消息，不能从历史事件中恢复 `"hello"`。

第 8 至第 12 课的示例也可分别复制到独立 `.sol` 文件，在 Remix VM 中用 Solidity 0.8.20 或更高的 0.8.x 编译器验证。示例用于理解机制，生产合约还需要完整测试和安全审查。

---



## 13. 错误处理与自定义错误

**要会什么**

- `require(条件, "原因")`：输入和业务前提。
- `assert(条件)`：永远不该被打破的内部关系。
- `revert`：想主动失败时使用。
- Solidity 0.8.4 起可以定义 `error InsufficientBalance(uint256 requested, uint256 available);`，再用 `revert InsufficientBalance(amount, balance);`。自定义错误比长字符串更省 Gas，也更方便链下解码。
- `try` / `catch` 只能包外部调用。用来处理“对方合约失败时我该怎么办”，不是用来包自己合约里的 `require`。

**注意**

- 函数开头就把非法输入挡掉，后面才改状态。
- 不要用 `assert` 检查用户输入。用户余额不足是正常业务失败，用 `require` 或自定义 error。
- 外部调用失败时，先看自己要不要回滚整笔交易。有时应该 `catch` 后继续，有时应该让它整体失败。

**示例：业务错误、自定义错误与外部调用失败**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract NumberSourceDemo {
    function doubleValue(uint256 value) external pure returns (uint256) {
        require(value > 0, "zero value");
        return value * 2;
    }
}

contract ErrorHandlingDemo {
    error InsufficientCredits(uint256 requested, uint256 available);

    uint256 public constant LIMIT = 100;
    uint256 public credits;

    function add(uint256 amount) external {
        require(amount > 0 && amount <= LIMIT - credits, "invalid amount");
        credits += amount;
        assert(credits <= LIMIT);
    }

    function spend(uint256 amount) external {
        if (amount > credits) {
            revert InsufficientCredits(amount, credits);
        }
        credits -= amount;
    }

    function tryDouble(NumberSourceDemo source, uint256 value)
        external view returns (bool ok, uint256 result)
    {
        try source.doubleValue(value) returns (uint256 doubled) {
            return (true, doubled);
        } catch {
            return (false, 0);
        }
    }
}
```

在 Remix 中先部署 `NumberSourceDemo`，再部署 `ErrorHandlingDemo`。`add(40)` 后 `credits()` 为 `40`；`spend(50)` 会显示自定义错误，传入的请求量为 `50`、可用量为 `40`。把数据源合约地址传给 `tryDouble`：输入 `3` 返回 `(true, 6)`，输入 `0` 因对方 `require` 失败而返回 `(false, 0)`。`assert` 用于检查合约自身上限不变量，不用于检查用户输入。这里 credits 只是演示用计数，不对应真实资金或账户余额。

---



## 14. 合约间调用

**要会什么**

- 高层调用：定义 `interface`，再按接口调用。类型清楚，返回值也清楚，优先用这种方式。
- 底层调用三种：
  - `call`：在对方合约里执行，`msg.sender` 变成你这个合约，可以附带以太币。
  - `delegatecall`：在自己的合约上下文里执行对方的代码，改的是自己的 storage。代理升级靠它。
  - `staticcall`：只读调用，对方如果改状态会失败。
- `new Token(...)` 由当前合约部署一个新合约。地址在部署时才能确定。`create2` 可以提前算出地址。

**注意**

- `call` 和 `delegatecall` 的返回值必须检查。失败时底层调用默认不回滚你的交易。
- `delegatecall` 要求两边的存储布局一致。槽位错一位，数据就会写坏。
- 外部调用是重入的入口。调用之前先把余额清掉或加上重入锁。
- 不要把不受信任的地址直接 `delegatecall`。那等于把合约的存储控制权交出去。

**示例：接口调用、staticcall 与工厂部署**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

interface IValueReaderDemo {
    function getValue() external view returns (uint256);
}

contract ValueStoreDemo is IValueReaderDemo {
    uint256 private value;

    constructor(uint256 initialValue) {
        value = initialValue;
    }

    function getValue() external view returns (uint256) {
        return value;
    }
}

contract ValueCallerDemo {
    function readTyped(address target) external view returns (uint256) {
        require(target.code.length > 0, "not a contract");
        return IValueReaderDemo(target).getValue();
    }

    function readStatic(address target) external view returns (uint256) {
        require(target.code.length > 0, "not a contract");
        (bool success, bytes memory data) = target.staticcall(
            abi.encodeCall(IValueReaderDemo.getValue, ())
        );
        require(success && data.length == 32, "read failed");
        return abi.decode(data, (uint256));
    }

    function createStore(uint256 initialValue) external returns (address) {
        ValueStoreDemo store = new ValueStoreDemo(initialValue);
        return address(store);
    }
}
```

部署 `ValueCallerDemo`，调用 `createStore(42)`，从交易输出中取得新合约地址；用此地址分别调用 `readTyped` 和 `readStatic`，都应返回 `42`。也可以在 Remix 中单独部署 `ValueStoreDemo(42)`，再用它的地址读取。传入外部账户地址会因代码长度检查回滚；目标合约即使有代码也未必实现接口，调用失败同样会回滚。

高层接口调用更适合固定的类型化接口；`staticcall` 需要手动检查成功标志并解码返回数据。本例不用 `delegatecall`，因为它会执行目标代码并修改调用方存储，只应在明确存储布局与升级权限的场景使用。

---



## 15. 合约安全基础

这一课把前面的坑收成攻击场景。写涉及钱的合约之前，按这张表过一遍。


| 漏洞   | 会发生什么                           | 怎么防                                           |
| ---- | ------------------------------- | --------------------------------------------- |
| 重入   | 转账时对方合约再调回来，余额还没扣，钱被提走多次        | 先改账再转账（CEI）；再加 `nonReentrant`                 |
| 整数溢出 | 0.8 之前 `uint8` 从 255 加 1 变成 0   | 使用 0.8 及以上；只在确认不会溢出的循环里用 `unchecked`          |
| 越权   | 谁都能 `mint`、换 owner、提款           | `onlyOwner` 或 OpenZeppelin `Ownable`；敏感函数逐个检查 |
| 拒绝服务 | 一次循环给所有人转账，其中一个人回滚或 Gas 不够，整批失败 | 改成拉取模式：每个人自己来取，而不是合约推给所有人                     |
| 抢跑   | 交易先出现在内存池里，别人看到后插队              | 提交哈希再揭示明文（commit-reveal）；或走私有交易通道             |


**注意**

- The DAO 事件就是先转账、后扣余额。Parity 多签事件则是权限初始化可以被外人调用。两个都是“检查漏了或顺序错了”。
- 推送模式 `for` 循环里 `transfer` 既贵又容易被一个地址卡死。拉取模式让用户调用 `withdraw` 取自己的钱。
- 部署前至少做一次：权限、重入、外部调用返回值、有没有无界循环。

**示例：拉取式提现、CEI 与重入锁**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract SafePaymentsDemo {
    mapping(address => uint256) public pending;
    bool private locked;

    function fund(address recipient) external payable {
        require(recipient != address(0), "zero recipient");
        require(msg.value > 0, "zero value");
        pending[recipient] += msg.value;
    }

    function withdraw() external {
        require(!locked, "reentrant call");
        uint256 amount = pending[msg.sender];
        require(amount > 0, "nothing to withdraw");
        locked = true;
        pending[msg.sender] = 0;
        (bool success, ) = payable(msg.sender).call{value: amount}("");
        require(success, "transfer failed");
        locked = false;
    }
}

contract ReentryProbeDemo {
    SafePaymentsDemo public immutable payments;
    bool public reentrySucceeded;

    constructor(SafePaymentsDemo target) {
        payments = target;
    }

    function collect() external {
        payments.withdraw();
    }

    receive() external payable {
        (bool success, ) = address(payments).call(
            abi.encodeCall(SafePaymentsDemo.withdraw, ())
        );
        reentrySucceeded = success;
    }
}
```

在 Remix VM 中先部署 `SafePaymentsDemo`，再以其地址为参数部署 `ReentryProbeDemo`。在 Value 输入 `1 ether`，调用支付合约的 `fund(探针合约地址)`；然后把 Value 重置为 `0`，调用探针的 `collect()`。探针收到钱时尝试再次调用 `withdraw()`，但 `reentrySucceeded()` 为 `false`；`pending(探针地址)` 为 `0`，钱只付出一次。直接给普通账户充值并调用 `withdraw()` 也能正常领款。

探针使用底层 `call` 观察失败而不回滚外层领取；真实攻击合约可能选择回滚收款，此时原交易会整体回滚，记录也不会被错误清零。这里没有遍历收款人，避免一个地址阻塞所有人的付款。示例锁用于展示原理，生产项目优先使用经测试的 `ReentrancyGuard` 并编写攻击测试。

---



## 16. Gas 优化

**要会什么**

费用等于「用了多少 Gas」乘「Gas 价格」。同样的功能，存储写得越少越便宜。一次新的 `SSTORE` 大约抵得上几千次加法。

这门课强调的六类做法：

1. 变量打包：小类型连续声明，让它们共用一个槽。
2. 局部存储指针：同一个 storage struct 不要每改一个字段就重新算一次位置。
3. 类型选择：要打包用小整数；单独存放、频繁运算用 `uint256`，避免来回转换。
4. 函数：外部入口用 `external`；短路时把便宜条件放前面。
5. 能用事件通知链下的历史，不要全部写进 storage。事件比存一份历史便宜一个数量级。
6. 批量处理放进一笔交易，分摊基础交易开销。确定不会溢出的计数循环可以用 `unchecked`，每次运算大约省 20 Gas。

**注意**

- 先保证正确和安全，再省 Gas。`unchecked` 用错会重新引入溢出。
- mapping 适合按键查找；数组适合短列表和需要遍历的场景。不要为了省一次存储把安全检查删掉。
- 用 Hardhat Gas Reporter 或 Foundry gas snapshot 看数字，不要凭感觉改。

**示例：变量打包、缓存 storage 与受控 unchecked**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract GasDemo {
    uint128 public packedAmount;
    uint64 public packedCount;
    bool public packedEnabled;
    uint256 public separateValue;

    uint256[] public values;

    function setPacked(uint128 amount, uint64 count, bool enabled) external {
        packedAmount = amount;
        packedCount = count;
        packedEnabled = enabled;
    }

    function addValues(uint256[] calldata newValues) external {
        require(newValues.length <= 50, "too many values");
        for (uint256 i = 0; i < newValues.length; i++) {
            values.push(newValues[i]);
        }
    }

    function sumCached() external view returns (uint256 total) {
        uint256 length = values.length;
        for (uint256 i = 0; i < length; i++) {
            total += values[i];
        }
    }

    function sumUnchecked() external view returns (uint256 total) {
        uint256 length = values.length;
        for (uint256 i = 0; i < length; ) {
            total += values[i];
            unchecked {
                ++i;
            }
        }
    }
}
```

部署后调用 `setPacked` 和直接写入状态的函数，在 Remix 的交易详情中比较 Gas；`packedAmount`、`packedCount`、`packedEnabled` 是连续的小字段，编译器可能把它们放进更少的 slot。`separateValue` 没有参与打包，用来对比布局时观察编译器的 storage layout。

调用 `addValues([1,2,3])` 后，`sumCached()` 和 `sumUnchecked()` 都返回 `6`。数组长度被限制为 50，循环边界明确，因此计数器在 `unchecked` 中不会超过 uint256；如果数组长度或循环逻辑可能失控，就不能使用 unchecked。是否真的省 Gas 应以 Hardhat Gas Reporter 或 Foundry snapshot 的结果为准，不能只凭代码外观判断。

第 13 至第 16 课的示例均可分别复制到独立 `.sol` 文件，在 Remix VM 中使用 Solidity 0.8.20 或更高的 0.8.x 编译器验证。安全示例只用于理解攻击面和防护顺序，生产代码应优先采用 OpenZeppelin 实现并配合自动化测试。

---



## 17. 设计模式

**要会什么**


| 模式   | 解决什么        | 典型写法                                 |
| ---- | ----------- | ------------------------------------ |
| 访问控制 | 谁能执行敏感操作    | `Ownable`，或多角色                       |
| 提现模式 | 转账失败或重入拖垮合约 | 先记账，用户自己拉取；配合 CEI                    |
| 状态机  | 众筹、拍卖有阶段    | `enum` 加每个函数开头的状态检查                  |
| 代理   | 逻辑要升级、地址不能变 | 代理合约保存数据，`delegatecall` 到逻辑合约        |
| 工厂   | 批量部署同类合约    | 一个工厂合约负责 `new`，并记录所有子合约地址            |
| 紧急停止 | 出事后立刻停掉入口   | `paused` 标志，关键函数先 `require(!paused)` |


**注意**

- 一个合约通常同时用好几个模式。例如市场合约会同时有访问控制、重入锁、紧急停止和拉取提现。
- 代理的存储槽不能随便插入新变量，否则升级后旧数据会错位。
- 优先用 OpenZeppelin 里已经审计过的实现，不要自己重写一套所有权。

---



## 18. NFT 市场

**要会什么**

- ERC20 的每个单位都一样。ERC721 的每一枚代币有独立 `tokenId`，`ownerOf(tokenId)` 查的是这一枚属于谁。
- 核心接口包括 `balanceOf`、`ownerOf`、`transferFrom`、`approve`，再加元数据扩展 `name`、`symbol`、`tokenURI`。
- 铸造流程是：递增计数器，把 `tokenId` 记到接收者名下，发出 `Transfer`。
- 市场上架用一个 `Listing` 结构：卖家、NFT 合约、tokenId、价格、是否还在售。
- 购买函数是 `payable`：检查钱够不够、listing 是否有效，把 NFT 转给买家，把钱结算给卖家，最后关掉 listing。
- ERC2981 用 `royaltyInfo(tokenId, salePrice)` 计算创作者版税。拍卖则另有起拍价、最高出价者和结束时间。

**注意**

- 购买和拍卖出价都会收以太币，并且要调用外部 NFT 合约，必须防重入。
- 下架、购买、拍卖结束都要检查“当前调用者是不是卖家或最高出价者”，以及“状态是不是还在进行”。
- 退还上一个出价者时，优先记录“可领取金额”，让对方自己取，避免在出价函数里给外部地址转账。

---



## 贯穿全课的注意要点

这些点会在多课里重复出现。写作业时先对照这里。

**数据放哪**

- 跨交易还要在的，才写 storage。函数算完就返回的，放 memory。外部只读参数用 calldata。
- mapping 不能遍历。要“重置所有人”或“列出所有人”，就额外维护一个数组，并用一个 `bool` 防止同一个人被推进数组两次。
- `delete` 数组元素不会缩短长度。`delete` mapping 的一个键只会恢复默认值。

**钱和权限**

- 收以太币的函数加 `payable`。转出时接收方必须是 `address payable`。
- 权限用 `msg.sender`，不用 `tx.origin`。
- 改余额、删除订单、标记已领取，都要发生在外部转账之前。
- `transfer` 只给对方 2300 Gas。对方是合约时可能收不了款。生产代码更常见的是 `call` 并检查返回值。
- 谁都能调用的函数，默认就是公开接口。铸造、提款、改管理员必须有 modifier。

**类型和运算**

- 0.8 及以上才有默认溢出保护。显式缩小类型之前先做范围检查。
- 字符串比较走哈希，不要用 `==`。
- `view` 不能改状态，`pure` 连状态和区块变量都不能读。

**Gas 和可用性**

- 链上循环必须有上限。给全体用户推送资金，会又贵又容易被拒绝服务。
- 循环里缓存 `length` 和 storage 读数。小字段打包后能少占槽。
- 事件用来通知外界，不用来保存合约还要读的账。

---



## 和已写作业的对应关系


| 已写代码                             | 用到的课                                                                                                 |
| -------------------------------- | ---------------------------------------------------------------------------------------------------- |
| `homework01/checkMark.sol` 的投票合约 | 第 5 课的 mapping 加数组；第 6 课的函数可见性                                                                       |
| 反转字符串                            | 第 2 课的 memory，第 3 课的 `bytes` 和 `string`                                                              |
| 合并有序数组、二分查找                      | 第 4 课的 memory 数组和第 7 课的循环；二分查找用下标把区间缩小一半                                                             |
| `homework02/BeggingContract.sol` | 第 6 课的 `payable`，第 8 课的 `msg.value`、`address(this).balance` 和 `transfer`，第 1 课的 `owner = msg.sender` |


讨饭合约目前用 `transfer`，这是作业指定的写法。学到第 8 课和第 15 课时，再对照理解：如果接收方是合约，2300 Gas 可能不够，那时应改成检查返回值的 `call`，并且仍然先记账、后转账。

---



## 总结

这门课可以分成四段。

第一段是第 1 课到第 3 课，解决“合约是什么、数据放在哪、类型怎么选”。这里最重要的结论是：storage 又贵又永久，memory 只活在一次调用里，calldata 只读而且更省；0.8 会拦住整数溢出；字符串和地址都有自己的比较、转换规则。

第二段是第 4 课到第 8 课，解决“数据怎么组织、函数怎么暴露、流程怎么写、钱怎么转”。数组负责有序列表，mapping 负责按键查找，struct 负责把一条记录的字段放在一起。函数必须写明谁能调用、会不会改状态、能不能收款。转账前先改自己的账，权限看 `msg.sender`。

第三段是第 9 课到第 14 课，开始写能和其他合约配合的代码。ERC20 把余额和授权拆开，继承和库用来复用，事件用来通知链下，自定义错误用来表达失败原因，接口和 `call` 用来调用外部合约。`delegatecall` 只在自己完全理解存储布局时使用。

第四段是第 15 课到第 18 课，把安全和结构补上。重入、越权、无界循环、抢跑是涉及资金时的四类常见事故。Gas 优化建立在功能已经正确的前提上。设计模式是这些防护的固定形状。NFT 市场是综合练习：ERC721 记录所有权，市场合约管上架和成交，购买路径上同时要用付款、权限、状态和重入防护。

复习时不要按课件页码回看。按一条交易把整门课串起来：谁调用（`msg.sender`）→ 参数放在哪（calldata / memory）→ 检查是否通过（require、modifier、状态机）→ 改哪些 storage（mapping、struct、数组）→ 要不要对外转账或调用别的合约 → 发出什么事件 → 失败时用什么错误回滚。这条链路上的每一步都能在上面的某一课里找到对应规则。

---



## Web3 还需要学什么

这 18 课学完，你能在 Remix 里把合约写对。离「能做一个别人用得上的 Web3 应用」还差一整段：交易到底怎么上链、合约怎么测试和部署、页面怎么跟钱包说话。这些课件没有讲，但比再学一门新语言更要紧。

我的判断是：先把以太坊这一条链走通，再决定要不要看 Solana、Rust 或零知识。现在同时开第二条链，合约基础会停在作业水平。

你的课表里，阶段三就是这件事，顺序已经排好了：Hardhat3 环境 → Hardhat3 测试 → Foundry → Ethers.js 与 MetaMask → 众筹、多签、NFT 盲盒。跟着这个顺序走，不必再另选一套工具。下面第 2 步到第 4 步是在解释这 7 节课各自补上了什么，不是让你跳过课表自己重排。

阶段三依赖前面的合约基本功。投票、反转字符串、合并数组、二分查找、讨饭合约还不能独立写出来之前，先不要看第 19 节。平台上即使标了已学完，也建议等前面讲清楚再回头看这 7 节。

### 合约课之后，按这个顺序补

**1. 搞清一笔交易在链上发生了什么**

课件从「合约代码」讲起，默认你已经知道账户和交易。其实要补这几件事：外部账户和合约账户的区别、nonce、Gas limit 和 Gas price 各自管什么、一笔交易从签名到进区块的过程、主网和测试网（现在练习用 Sepolia）的差别。wei、gwei、ether 的换算也要熟，否则 `msg.value` 只是一个大整数。

练法：用 Remix 或小狐狸钱包往测试网部署讨饭合约，在区块浏览器里打开那笔交易，对着 input data、logs、value、gas used 逐项认一遍。认不出这些字段，后面的部署和调试都会靠猜。

**2. 离开 Remix，先 Hardhat，再 Foundry**

阶段三的第 19、20 节用 Hardhat3 把「编译、部署、单元测试」从网页里搬到项目里。第 21 节再用 Foundry 做同一件事。两套都要学，因为课是这么排的：Hardhat 带你认识脚本和 JavaScript 测试，Foundry 的测试直接用 Solidity 写，读起来更接近合约本身。

学的时候不要两套同时开工。第 19、20 节只跟 Hardhat，用讨饭合约写三组测试：`donate` 之后 `getDonation` 对不对、别人调用 `withdraw` 会不会失败、所有者提走之后余额是不是 0。第 21 节再用 Foundry 把这三组重写一遍。重写是在对照工具，不是在学新业务。

**3. 把 OpenZeppelin 当成标准库来用**

课件提到了 Ownable 和 ERC20，但作业都是手写。下一步不是背接口，而是会导入现成合约：所有权用 `Ownable`，代币继承 `ERC20`，暂停用 `Pausable`，重入锁用 `ReentrancyGuard`。先读它的 `withdraw`、`_mint`、`nonReentrant` 实现，再在自己的合约里继承。自己重写一套所有权和代币，容易把第 15 课的漏洞又写回来。

**4. 做一条完整链路：合约、钱包、页面**

只写合约，用户碰不到。最小闭环就用讨饭合约：

- 合约部署到 Sepolia
- 页面用钱包连上
- 读 `getDonation`
- 调 `donate` 并附带以太币
- 从 `Donation` 这类事件里看到「谁捐了」

这一步对应第 22 节，库就用 Ethers.js 配 MetaMask，不要再并行学 viem。钱包连接用浏览器注入的 MetaMask 即可，先不要上一套账户抽象。

这一步会逼你理解 ABI：页面不是在「调用函数名」这么简单，它是按 ABI 编码参数、发送交易、再解码返回值和日志。

**5. 用闯关把安全课变成手感**

第 15 课的重入、`tx.origin`、`delegatecall` 看懂和写得出来是两回事。课件过完后去打 Ethernaut，优先做 Fallback、Fallout、Coin Flip、Telephone、Delegation、Reentrancy 这几关。每关先自己改合约，再看题解。Damn Vulnerable DeFi 放到能读懂一个简单兑换或借贷合约之后，现在做会卡住。

### 再往后，两条路选一条深入

合约方向：把 Uniswap V2 的恒定乘积读懂（`x * y = k` 在加减流动性时池子两边怎么变），再看一个借贷协议怎么记账。预言机只需要先知道「合约自己看不到链下价格，价格是喂进来的」。代理升级课里已经有了，等存储布局真的写错过一次再深入，不要一开始就做可升级合约。

应用方向：把事件同步做扎实。页面刷新后来查历史，不能只靠「当时那次调用的返回值」。先自己用 `eth_getLogs` 把讨饭合约的捐赠记录拉出来，再考虑 The Graph。索引服务是这层需求变大之后的工具，不是入门必修。

### 现在先不要学的

- 第二条链。Solidity 还没配上测试和一次测试网部署之前，学 Rust 或 Move 只会多一套语法。
- 手写汇编、Yul、自研代币经济模型。
- 一上来就做账户抽象、跨链桥、零知识电路。那些是产品形态，不是基本功。
- 为了省 Gas 把合约改到难读。第 16 课的优化，等 Foundry 能打出 Gas 数字再做。



### 建议的时间切法


| 阶段    | 做什么                                           | 做到就算过关                      |
| ----- | --------------------------------------------- | --------------------------- |
| 现在    | 把前面的合约课和现有作业写完                                | 讨饭、投票、数组题能在 Remix 里独立写出来    |
| 阶段三前半 | 第 19–21 节：Hardhat3，再用 Foundry 重写同一批测试         | 讨饭合约在两个框架里都能测过，并部署到 Sepolia |
| 阶段三后半 | 第 22–25 节：Ethers.js 连 MetaMask，再做众筹、多签、NFT 盲盒 | 能从页面捐一笔；三个实战能讲清钱存在哪、谁有权动它   |
| 接着    | Ethernaut 前几关                                 | 能口头讲清重入为什么要先改账              |
| 最后再分叉 | 读 Uniswap V2，或把事件索引做进页面                       | 能跟着一笔真实交易把代码路径走完            |


进第 23 节之前，前面这些要对得上，否则实战会变成照着视频敲：


| 阶段三                     | 开看前要已经会                                           |
| ----------------------- | ------------------------------------------------- |
| 19–20 Hardhat3          | 状态变量、`require`、`payable`、事件。否则测试里断言的对象你还不认识       |
| 21 Foundry              | 同一批合约能在 Hardhat 里跑通。Foundry 换的是工具，不是新语法           |
| 22 Ethers.js 与 MetaMask | ABI、`view` 和发交易的区别、事件日志。页面读的是 `view`，捐款是发交易       |
| 23 众筹                   | `enum` 状态机、`block.timestamp`、`onlyOwner`、先改账再转账   |
| 24 多签                   | 数组加 mapping、达到门槛后再用 `call` 执行外部调用                 |
| 25 NFT 盲盒               | ERC721 的 `tokenId` 所有权，以及链上随机数可以被抢跑，所以盲盒通常要先承诺再揭示 |


Web3 岗位常见的分工也是这样：写合约的人要会测试和安全；做应用的人要会钱包、ABI 和事件。两个方向都建立在「一笔交易从签名到改 storage、再到打出日志」这条链上。课件把中间的 Solidity 补上了，头尾还要你自己接上。

## 基础二：Solidity智能合约基础

学习课件：[https://github.com/MetaNodeAcademy/solidity_lesson；](https://github.com/MetaNodeAcademy/solidity_lesson；)

学习视频：先登录小鹅通学习平台，然后根据课件中的章节目录，点击对应知识点链接即可跳转观看配套视频

作业提交方式：在个人 GitHub 账号下创建仓库，将作业代码提交至仓库，并提交对应的 GitHub 仓库链接给导师检查

总体学习时间：3周

### 1）第一阶段：基础入门

掌握Solidity核心语法和基础编程能力。包括智能合约概念、开发环境搭建、EVM存储机制、数据类型、数组、映射、结构体、函数、修饰符、控制流语句等基础知识，并完成简单代币合约项目。
课件范围:  lesson1.1~lesson5.1
使用在线的IDE: Remix

### 2）第二阶段：进阶编程

深入学习合约高级特性和安全开发。包括合约继承、库合约、事件、错误处理、合约间调用、智能合约安全、Gas优化技巧、设计模式等进阶主题，并完成NFT市场实战项目。
课件范围: lesson6.1~lesson10.1
使用在线的IDE: Remix

### 3）第三阶段：开发框架与实战

掌握专业开发工具和实战项目开发。学习Hardhat和Foundry两大开发框架、单元测试、Ethers.js、MetaMask集成等开发技能，通过众筹平台、多签钱包、NFT盲盒等完整项目提升实战能力。
这两年新出现的项目，大多优先选择 Foundry 作为开发与测试工具链。

课件范围: lesson11.1~lesson14.3
使用的开发环境: VSCode编辑器 + Solidity插件 + Hardhat&Foundry工具

#### 官方文档

Foundry官方文档：[https://getfoundry.sh/introduction/getting-started；](https://getfoundry.sh/introduction/getting-started；)
Hardhat官方文档：[https://v2.hardhat.org/docs；](https://v2.hardhat.org/docs；)

### 4）第四阶段：完成简单作业

[https://github.com/MetaNodeAcademy/LearningRoadmap/blob/main/contract/homework01.md](https://github.com/MetaNodeAcademy/LearningRoadmap/blob/main/contract/homework01.md)
[https://github.com/MetaNodeAcademy/LearningRoadmap/blob/main/contract/homework02.md](https://github.com/MetaNodeAcademy/LearningRoadmap/blob/main/contract/homework02.md)

### 5）掌握合约升级原理

[https://github.com/MetaNodeAcademy/Base2-Solidity/tree/main/3-solidity-coding/3.2%E5%8F%AF%E5%8D%87%E7%BA%A7%E5%90%88%E7%BA%A6（文档）](https://github.com/MetaNodeAcademy/Base2-Solidity/tree/main/3-solidity-coding/3.2%E5%8F%AF%E5%8D%87%E7%BA%A7%E5%90%88%E7%BA%A6（文档）)

[https://meeting.tencent.com/wework/cloud-record/share?id=8acee3d0-2fc2-48f3-a702-b33e9906bd64&record_type=2&hide_more_btn=true&from=qywx（视频）](https://meeting.tencent.com/wework/cloud-record/share?id=8acee3d0-2fc2-48f3-a702-b33e9906bd64&record_type=2&hide_more_btn=true&from=qywx（视频）)

访问密码：0ZEX

### 6）跨链与预言机

[https://github.com/MetaNodeAcademy/Base2-Solidity/blob/main/3-solidity-coding/3.4%E8%B7%A8%E9%93%BE/01.CCIP.md](https://github.com/MetaNodeAcademy/Base2-Solidity/blob/main/3-solidity-coding/3.4%E8%B7%A8%E9%93%BE/01.CCIP.md)

[https://github.com/MetaNodeAcademy/Base2-Solidity/tree/main/3-solidity-coding/3.3%E9%A2%84%E8%A8%80%E6%9C%BA%E9%9B%86%E6%88%90](https://github.com/MetaNodeAcademy/Base2-Solidity/tree/main/3-solidity-coding/3.3%E9%A2%84%E8%A8%80%E6%9C%BA%E9%9B%86%E6%88%90)

### 7）完成任务3的作业

根据个人技术背景和使用习惯，可自行选择使用 Foundry 或 Hardhat 作为开发框架完成作业。

[https://github.com/MetaNodeAcademy/LearningRoadmap/blob/main/contract/homework03.md](https://github.com/MetaNodeAcademy/LearningRoadmap/blob/main/contract/homework03.md)

参考代码  ：[https://github.com/lc3091/hardhatV3Nft.git](https://github.com/lc3091/hardhatV3Nft.git)

### 8）推荐阅读其他文档

OpenZeppelin: [https://docs.openzeppelin.com/contracts/5.x/，建议先通读官方文档中的介绍部分，了解](https://docs.openzeppelin.com/contracts/5.x/，建议先通读官方文档中的介绍部分，了解) OpenZeppelin 提供了哪些常用合约、库和功能模块，建立整体认知。实际开发过程中，当需要使用某个功能时，再深入查阅对应的 API 和实现细节，无需一开始就记忆所有内容。

Solidity语言Cheatsheet：[https://docs.soliditylang.org/en/v0.8.28/cheatsheet.html](https://docs.soliditylang.org/en/v0.8.28/cheatsheet.html)
Remix使用方法：[https://github.com/AmazingAng/WTF-Solidity/tree/main/Topics/Tools/TOOL01_Remix](https://github.com/AmazingAng/WTF-Solidity/tree/main/Topics/Tools/TOOL01_Remix)

WTF-Solidity: [https://github.com/WTFAcademy/WTF-Solidity](https://github.com/WTFAcademy/WTF-Solidity)