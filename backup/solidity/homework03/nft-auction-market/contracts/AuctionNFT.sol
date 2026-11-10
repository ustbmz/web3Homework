// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {ERC721} from "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

/// @title AuctionNFT
/// @notice 拍卖市场上架用的 ERC721。铸造和转移是两件分开的事：
///         铸造是“凭空产生一枚新的 NFT”，转移是“把已经存在的一枚交给别人”。
///         转移不需要自己再写一遍，ERC721 已经提供 transferFrom / safeTransferFrom。
contract AuctionNFT is ERC721, Ownable {
    uint256 public nextTokenId;
    mapping(uint256 tokenId => string uri) private _tokenUris;

    constructor(address initialOwner) ERC721("Homework Auction NFT", "HANFT") Ownable(initialOwner) {}

    /// 只有合约所有者能铸造。作业里的“转移”用继承来的 safeTransferFrom，不在这里重复实现。
    /// @param to  第一任持有人。拍卖前，这个地址要先把 NFT 授权给拍卖代理合约。
    /// @param uri 这枚 NFT 的元数据地址，例如 ipfs://...
    /// @return tokenId 新 NFT 的编号，从 0 开始递增。
    function mint(address to, string calldata uri) external onlyOwner returns (uint256 tokenId) {
        tokenId = nextTokenId++;
        _safeMint(to, tokenId);
        _tokenUris[tokenId] = uri;
    }

    /// ERC721 要求存在的 token 才能返回 URI。_requireOwned 会在编号不存在时回滚。
    function tokenURI(uint256 tokenId) public view override returns (string memory) {
        _requireOwned(tokenId);
        return _tokenUris[tokenId];
    }
}
