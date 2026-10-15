import { expect } from "chai";
import { network } from "hardhat";

const { ethers, networkHelpers } = await network.create();

const HOUR = 3600;
const ETH_PRICE = 2000n * 10n ** 8n;
const USDC_PRICE = 1n * 10n ** 8n;

async function deployFixture() {
  const [owner, seller, seller2, bidder1, bidder2, treasury] = await ethers.getSigners();

  const nft = await ethers.deployContract("AuctionNFT", [owner.address]);
  const usdc = await ethers.deployContract("MockERC20", ["Mock USDC", "USDC", 6]);
  const ethFeed = await ethers.deployContract("MockV3Aggregator", [8, ETH_PRICE]);
  const usdcFeed = await ethers.deployContract("MockV3Aggregator", [8, USDC_PRICE]);

  const implementation = await ethers.deployContract("NftAuction");
  const initData = implementation.interface.encodeFunctionData("initialize", [
    owner.address,
    await ethFeed.getAddress(),
    treasury.address,
    3n * 60n * 60n,
  ]);
  const proxy = await ethers.deployContract("AuctionProxy", [await implementation.getAddress(), initData]);
  const auction = await ethers.getContractAt("NftAuction", await proxy.getAddress());
  await auction.setPriceFeed(await usdc.getAddress(), await usdcFeed.getAddress());

  const mintAmount = ethers.parseUnits("1000", 6);
  await usdc.mint(bidder1.address, mintAmount);
  await usdc.mint(bidder2.address, mintAmount);
  await usdc.connect(bidder1).approve(await auction.getAddress(), ethers.MaxUint256);
  await usdc.connect(bidder2).approve(await auction.getAddress(), ethers.MaxUint256);

  return {
    owner,
    seller,
    seller2,
    bidder1,
    bidder2,
    treasury,
    nft,
    usdc,
    ethFeed,
    usdcFeed,
    implementation,
    proxy,
    auction,
  };
}

async function startAuction(
  nft: Awaited<ReturnType<typeof deployFixture>>["nft"],
  auction: Awaited<ReturnType<typeof deployFixture>>["auction"],
  seller: Awaited<ReturnType<typeof deployFixture>>["seller"],
) {
  const tokenId = await nft.nextTokenId();
  await nft.mint(seller.address, `ipfs://homework03/${tokenId}`);
  await nft.connect(seller).approve(await auction.getAddress(), tokenId);
  const auctionId = await auction.nextAuctionId();
  await auction.connect(seller).createAuction(await nft.getAddress(), tokenId, HOUR);
  return { tokenId, auctionId };
}

describe("AuctionNFT", function () {
  it("所有者可以铸造，其他人不能铸造", async function () {
    const { nft, owner, seller } = await networkHelpers.loadFixture(deployFixture);

    await expect(nft.connect(seller).mint(seller.address, "ipfs://nope")).to.be.revertedWithCustomError(
      nft,
      "OwnableUnauthorizedAccount",
    );

    const tokenId = await nft.nextTokenId();
    await nft.mint(seller.address, "ipfs://one");
    expect(await nft.ownerOf(tokenId)).to.equal(seller.address);
    expect(await nft.tokenURI(tokenId)).to.equal("ipfs://one");
    expect(await nft.owner()).to.equal(owner.address);
  });

  it("持有人可以用 ERC721 的 transferFrom 把 NFT 转给别人", async function () {
    const { nft, seller, bidder1 } = await networkHelpers.loadFixture(deployFixture);

    const tokenId = await nft.nextTokenId();
    await nft.mint(seller.address, "ipfs://move");
    // 接收方是普通地址时用 transferFrom。拍卖合约收 NFT 时用的是 safeTransferFrom，
    // 因为合约必须回应 onERC721Received，否则转账会回滚。
    await nft.connect(seller).transferFrom(seller.address, bidder1.address, tokenId);

    expect(await nft.ownerOf(tokenId)).to.equal(bidder1.address);
  });
});

describe("报价和手续费", function () {
  it("0.01 ETH 在 2000 美元的价格下等于 20 美元，30 USDC 等于 30 美元", async function () {
    const { auction, usdc } = await networkHelpers.loadFixture(deployFixture);

    // 10^16 * 2000e8 * 10^10 / 10^18 = 20e18
    expect(await auction.quoteUsd(ethers.ZeroAddress, ethers.parseEther("0.01"))).to.equal(ethers.parseEther("20"));
    // 30e6 * 1e8 * 10^10 / 10^6 = 30e18
    expect(await auction.quoteUsd(await usdc.getAddress(), ethers.parseUnits("30", 6))).to.equal(
      ethers.parseEther("30"),
    );
  });

  it("手续费按美元金额分三档", async function () {
    const { auction } = await networkHelpers.loadFixture(deployFixture);

    expect(await auction.feeBps(ethers.parseEther("99"))).to.equal(100n);
    expect(await auction.feeBps(ethers.parseEther("100"))).to.equal(200n);
    expect(await auction.feeBps(ethers.parseEther("999"))).to.equal(200n);
    expect(await auction.feeBps(ethers.parseEther("1000"))).to.equal(300n);
  });

  it("价格超过允许的更新时间就不能报价", async function () {
    const { auction, ethFeed } = await networkHelpers.loadFixture(deployFixture);

    await ethFeed.setUpdatedAt(1);
    await expect(auction.quoteUsd(ethers.ZeroAddress, ethers.parseEther("0.01"))).to.be.revertedWithCustomError(
      auction,
      "StalePrice",
    );
  });

  it("价格小于等于 0 时回滚", async function () {
    const { auction, ethFeed } = await networkHelpers.loadFixture(deployFixture);

    await ethFeed.setAnswer(-1);
    await expect(auction.quoteUsd(ethers.ZeroAddress, ethers.parseEther("0.01"))).to.be.revertedWithCustomError(
      auction,
      "InvalidPrice",
    );
  });
});

describe("创建拍卖", function () {
  it("卖家授权后，NFT 转到拍卖代理，而不是实现合约", async function () {
    const { nft, auction, proxy, seller } = await networkHelpers.loadFixture(deployFixture);

    const { tokenId, auctionId } = await startAuction(nft, auction, seller);
    expect(await nft.ownerOf(tokenId)).to.equal(await proxy.getAddress());

    const stored = await auction.auctions(auctionId);
    expect(stored.seller).to.equal(seller.address);
    expect(stored.tokenId).to.equal(tokenId);
    expect(stored.settled).to.equal(false);
  });

  it("不是主人、没授权、时间太短，都不能创建", async function () {
    const { nft, auction, seller, bidder1 } = await networkHelpers.loadFixture(deployFixture);

    const tokenId = await nft.nextTokenId();
    await nft.mint(seller.address, "ipfs://locked");

    await expect(auction.connect(bidder1).createAuction(await nft.getAddress(), tokenId, HOUR))
      .to.be.revertedWithCustomError(auction, "NotNftOwner")
      .withArgs(bidder1.address, tokenId);

    await expect(
      auction.connect(seller).createAuction(await nft.getAddress(), tokenId, 59),
    ).to.be.revertedWithCustomError(auction, "DurationTooShort");

    await expect(
      auction.connect(seller).createAuction(await nft.getAddress(), tokenId, HOUR),
    ).to.be.revertedWithCustomError(nft, "ERC721InsufficientApproval");
  });
});

describe("出价", function () {
  it("美元更高的 ERC20 能赢过原始数量更大的 ETH，并退回 ETH", async function () {
    const { nft, auction, usdc, seller, bidder1, bidder2 } = await networkHelpers.loadFixture(deployFixture);
    const { auctionId } = await startAuction(nft, auction, seller);

    const ethBid = ethers.parseEther("0.0001"); // 原始数量 10^14，美元只有 0.2
    const usdcBid = ethers.parseUnits("1", 6); // 原始数量 10^6，美元是 1

    await auction.connect(bidder1).bid(auctionId, ethers.ZeroAddress, ethBid, { value: ethBid });
    expect(await auction.quoteUsd(ethers.ZeroAddress, ethBid)).to.equal(ethers.parseEther("0.2"));
    expect(await auction.quoteUsd(await usdc.getAddress(), usdcBid)).to.equal(ethers.parseEther("1"));

    await expect(auction.connect(bidder2).bid(auctionId, await usdc.getAddress(), usdcBid)).to.changeEtherBalance(
      ethers,
      bidder1,
      ethBid,
    );

    const stored = await auction.auctions(auctionId);
    expect(stored.highestBidder).to.equal(bidder2.address);
    expect(stored.highestBidAmount).to.equal(usdcBid);
    expect(await ethers.provider.getBalance(auction)).to.equal(0n);
  });

  it("更低的美元价格不能覆盖当前出价", async function () {
    const { nft, auction, usdc, seller, bidder1, bidder2 } = await networkHelpers.loadFixture(deployFixture);
    const { auctionId } = await startAuction(nft, auction, seller);

    const ethBid = ethers.parseEther("0.01");
    await auction.connect(bidder1).bid(auctionId, ethers.ZeroAddress, ethBid, { value: ethBid });

    const low = ethers.parseUnits("10", 6);
    await expect(auction.connect(bidder2).bid(auctionId, await usdc.getAddress(), low))
      .to.be.revertedWithCustomError(auction, "BidTooLow")
      .withArgs(ethers.parseEther("10"), ethers.parseEther("20"));
  });

  it("卖家不能给自己出价，ERC20 出价不能附带 ETH", async function () {
    const { nft, auction, usdc, seller, bidder1 } = await networkHelpers.loadFixture(deployFixture);
    const { auctionId } = await startAuction(nft, auction, seller);

    await expect(
      auction.connect(seller).bid(auctionId, ethers.ZeroAddress, ethers.parseEther("0.01"), {
        value: ethers.parseEther("0.01"),
      }),
    ).to.be.revertedWithCustomError(auction, "SellerCannotBid");

    await expect(
      auction.connect(bidder1).bid(auctionId, await usdc.getAddress(), ethers.parseUnits("1", 6), {
        value: 1n,
      }),
    ).to.be.revertedWithCustomError(auction, "UnexpectedEth");
  });

  it("结束后不能再出价，时间没到不能结算", async function () {
    const { nft, auction, seller, bidder1 } = await networkHelpers.loadFixture(deployFixture);
    const { auctionId } = await startAuction(nft, auction, seller);

    await expect(auction.connect(bidder1).endAuction(auctionId)).to.be.revertedWithCustomError(
      auction,
      "AuctionStillRunning",
    );

    await networkHelpers.time.increase(HOUR);
    const ethBid = ethers.parseEther("0.01");
    await expect(
      auction.connect(bidder1).bid(auctionId, ethers.ZeroAddress, ethBid, { value: ethBid }),
    ).to.be.revertedWithCustomError(auction, "AuctionFinished");
  });
});

describe("结算", function () {
  it("无人出价时把 NFT 还给卖家", async function () {
    const { nft, auction, seller } = await networkHelpers.loadFixture(deployFixture);
    const { tokenId, auctionId } = await startAuction(nft, auction, seller);

    await networkHelpers.time.increase(HOUR);
    await auction.endAuction(auctionId);

    expect(await nft.ownerOf(tokenId)).to.equal(seller.address);
    const stored = await auction.auctions(auctionId);
    expect(stored.settled).to.equal(true);
  });

  it("ERC20 成交后，NFT 给赢家，卖家拿 99%，手续费进收款地址", async function () {
    const { nft, auction, usdc, seller, bidder1, bidder2, treasury } = await networkHelpers.loadFixture(deployFixture);
    const { tokenId, auctionId } = await startAuction(nft, auction, seller);

    const ethBid = ethers.parseEther("0.01");
    await auction.connect(bidder1).bid(auctionId, ethers.ZeroAddress, ethBid, { value: ethBid });

    const usdcBid = ethers.parseUnits("30", 6);
    await auction.connect(bidder2).bid(auctionId, await usdc.getAddress(), usdcBid);

    await networkHelpers.time.increase(HOUR);
    await auction.endAuction(auctionId);

    const fee = (usdcBid * 100n) / 10_000n;
    expect(await nft.ownerOf(tokenId)).to.equal(bidder2.address);
    expect(await usdc.balanceOf(seller.address)).to.equal(usdcBid - fee);
    expect(await usdc.balanceOf(treasury.address)).to.equal(fee);
    expect(await usdc.balanceOf(auction)).to.equal(0n);
  });

  it("两场 ETH 拍卖分开记账，结束一场不会把另一场的钱转走", async function () {
    const { nft, auction, seller, seller2, bidder1, bidder2, treasury } =
      await networkHelpers.loadFixture(deployFixture);
    const first = await startAuction(nft, auction, seller);
    const second = await startAuction(nft, auction, seller2);

    const firstBid = ethers.parseEther("0.01");
    const secondBid = ethers.parseEther("0.02");
    await auction.connect(bidder1).bid(first.auctionId, ethers.ZeroAddress, firstBid, { value: firstBid });
    await auction.connect(bidder2).bid(second.auctionId, ethers.ZeroAddress, secondBid, { value: secondBid });

    await networkHelpers.time.increase(HOUR);
    const fee = (firstBid * 100n) / 10_000n;
    await expect(auction.endAuction(first.auctionId)).to.changeEtherBalances(
      ethers,
      [seller, treasury],
      [firstBid - fee, fee],
    );

    expect(await ethers.provider.getBalance(auction)).to.equal(secondBid);
    expect(await nft.ownerOf(second.tokenId)).to.equal(await auction.getAddress());
  });

  it("1000 美元以上收 3%：1 ETH 成交，卖家得 0.97，手续费 0.03", async function () {
    const { nft, auction, seller, bidder1, treasury } = await networkHelpers.loadFixture(deployFixture);
    const { auctionId } = await startAuction(nft, auction, seller);

    const bid = ethers.parseEther("1");
    expect(await auction.quoteUsd(ethers.ZeroAddress, bid)).to.equal(ethers.parseEther("2000"));
    expect(await auction.feeBps(ethers.parseEther("2000"))).to.equal(300n);

    await auction.connect(bidder1).bid(auctionId, ethers.ZeroAddress, bid, { value: bid });
    await networkHelpers.time.increase(HOUR);

    const fee = (bid * 300n) / 10_000n;
    await expect(auction.endAuction(auctionId)).to.changeEtherBalances(ethers, [seller, treasury], [bid - fee, fee]);
  });
});

describe("边界", function () {
  it("错误的管理参数和错误的出价会被拒绝", async function () {
    const { nft, auction, ethFeed, treasury, owner, seller, bidder1 } =
      await networkHelpers.loadFixture(deployFixture);

    await expect(auction.setPriceFeed(ethers.ZeroAddress, ethers.ZeroAddress)).to.be.revertedWithCustomError(
      auction,
      "ZeroAddress",
    );
    await expect(auction.setFeeRecipient(ethers.ZeroAddress)).to.be.revertedWithCustomError(auction, "ZeroAddress");
    await expect(auction.setMaxPriceAge(0)).to.be.revertedWithCustomError(auction, "ZeroAmount");
    await expect(auction.connect(bidder1).setFeeRecipient(bidder1.address)).to.be.revertedWithCustomError(
      auction,
      "OwnableUnauthorizedAccount",
    );

    const fresh = await ethers.deployContract("NftAuction");
    const badInit = fresh.interface.encodeFunctionData("initialize", [
      ethers.ZeroAddress,
      await ethFeed.getAddress(),
      treasury.address,
      3600,
    ]);
    await expect(
      ethers.deployContract("AuctionProxy", [await fresh.getAddress(), badInit]),
    ).to.be.revertedWithCustomError(fresh, "ZeroAddress");

    const { auctionId } = await startAuction(nft, auction, seller);
    await expect(
      auction.connect(bidder1).bid(auctionId, ethers.ZeroAddress, 1n, { value: 2n }),
    ).to.be.revertedWithCustomError(auction, "EthAmountMismatch");
    await expect(
      auction.connect(bidder1).bid(999, ethers.ZeroAddress, 1n, { value: 1n }),
    ).to.be.revertedWithCustomError(auction, "AuctionNotFound");
    await expect(auction.quoteUsd(bidder1.address, 1n)).to.be.revertedWithCustomError(auction, "PriceFeedMissing");
    await expect(auction.connect(seller).createAuction(ethers.ZeroAddress, 0, HOUR)).to.be.revertedWithCustomError(
      auction,
      "ZeroAddress",
    );

    await networkHelpers.time.increase(HOUR);
    await auction.endAuction(auctionId);
    await expect(auction.endAuction(auctionId)).to.be.revertedWithCustomError(auction, "AuctionAlreadySettled");

    const wideFeed = await ethers.deployContract("MockV3Aggregator", [19, 1n]);
    await auction.connect(owner).setPriceFeed(ethers.ZeroAddress, await wideFeed.getAddress());
    await expect(auction.quoteUsd(ethers.ZeroAddress, 1n)).to.be.revertedWithCustomError(auction, "InvalidPrice");
  });
});

describe("升级", function () {
  it("换成 V2 后旧拍卖还在，并且可以要求新出价至少高 10%", async function () {
    const { nft, auction, proxy, owner, seller, bidder1, bidder2 } = await networkHelpers.loadFixture(deployFixture);
    const { auctionId } = await startAuction(nft, auction, seller);

    const firstBid = ethers.parseEther("0.01");
    await auction.connect(bidder1).bid(auctionId, ethers.ZeroAddress, firstBid, { value: firstBid });
    expect(await auction.version()).to.equal("v1");

    const implementationV2 = await ethers.deployContract("NftAuctionV2");
    await expect(
      auction.connect(bidder1).upgradeToAndCall(await implementationV2.getAddress(), "0x"),
    ).to.be.revertedWithCustomError(auction, "OwnableUnauthorizedAccount");

    await auction.connect(owner).upgradeToAndCall(await implementationV2.getAddress(), "0x");
    const auctionV2 = await ethers.getContractAt("NftAuctionV2", await proxy.getAddress());

    expect(await auctionV2.version()).to.equal("v2");
    expect((await auctionV2.auctions(auctionId)).highestBidder).to.equal(bidder1.address);
    expect((await auctionV2.auctions(auctionId)).highestBidAmount).to.equal(firstBid);

    await auctionV2.setMinIncrementBps(1000);
    const shortBid = ethers.parseEther("0.0105");
    await expect(
      auctionV2.connect(bidder2).bid(auctionId, ethers.ZeroAddress, shortBid, { value: shortBid }),
    ).to.be.revertedWithCustomError(auctionV2, "BidTooLow");

    const enoughBid = ethers.parseEther("0.011");
    await expect(
      auctionV2.connect(bidder2).bid(auctionId, ethers.ZeroAddress, enoughBid, { value: enoughBid }),
    ).to.changeEtherBalance(ethers, bidder1, firstBid);

    expect((await auctionV2.auctions(auctionId)).highestBidAmount).to.equal(enoughBid);
  });

  it("加价比例超过 20% 会被拒绝，第一笔出价不受加价比例限制", async function () {
    const { nft, auction, proxy, owner, seller, bidder1 } = await networkHelpers.loadFixture(deployFixture);
    const implementationV2 = await ethers.deployContract("NftAuctionV2");
    await auction.connect(owner).upgradeToAndCall(await implementationV2.getAddress(), "0x");
    const auctionV2 = await ethers.getContractAt("NftAuctionV2", await proxy.getAddress());

    await expect(auctionV2.setMinIncrementBps(2_001)).to.be.revertedWithCustomError(auctionV2, "IncrementTooHigh");
    await auctionV2.setMinIncrementBps(1000);

    const { auctionId } = await startAuction(nft, auctionV2, seller);
    const bid = ethers.parseEther("0.01");
    await auctionV2.connect(bidder1).bid(auctionId, ethers.ZeroAddress, bid, { value: bid });
    expect((await auctionV2.auctions(auctionId)).highestBidUsd).to.equal(ethers.parseEther("20"));
  });

  it("实现合约不能被初始化，代理也不能初始化第二次", async function () {
    const { implementation, auction, owner, ethFeed, treasury } = await networkHelpers.loadFixture(deployFixture);

    await expect(
      implementation.initialize(owner.address, await ethFeed.getAddress(), treasury.address, 3600),
    ).to.be.revertedWithCustomError(implementation, "InvalidInitialization");

    await expect(
      auction.initialize(owner.address, await ethFeed.getAddress(), treasury.address, 3600),
    ).to.be.revertedWithCustomError(auction, "InvalidInitialization");
  });
});
