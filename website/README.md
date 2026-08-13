# Praxeon Website

Praxeon 官方网站（静态站），包含三个页面：

| 路径 | 页面 | 说明 |
|---|---|---|
| `index.html` | 首页 | 火星文明时钟 + 四元素 + 三阶段 |
| `brand/index.html` | 品牌手册 | 11 章节：文明/经济/身份/Agent/区块链/视觉 |
| `plan/index.html` | GDD 总案 | 17 章产品规格 + 图示（`plan/gdd.md`） |

## 技术

- 纯静态 HTML/CSS/JS，无构建步骤
- 火星图：NASA 影像（公共领域），WebP 响应式（640/1000/1600/2400）
- 品牌：纯文字 PRAXEON wordmark + Mars Signal 红（#E44B2D）

## 部署

```bash
# 任意静态服务器，如：
python3 -m http.server 8000
# 或 nginx 指向本目录
```

## 火星时钟

首页 hero 显示火星历（Darian 24 个月）实时时钟。纪元（Mars Epoch）常量在 `index.html` 的 `MARS_EPOCH`，正式启动时更新。
