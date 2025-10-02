# CPU性能测试排行榜

本目录包含CPU性能测试的结果和排行榜数据。

## 文件说明

### 主要文件
- `index.html`: 排行榜网页展示页面
- `all_cpu_results.json`: 包含所有CPU测试结果的完整数据
- `CNAME`: GitHub Pages自定义域名配置文件

### CSV文件夹 (csvs/)
- `all_cpus_single_core_ranking.csv`: 所有CPU单核性能排行榜
- `all_cpus_multi_core_ranking.csv`: 所有CPU多核性能排行榜
- `{CPU型号}_single_core.csv`: 特定CPU型号的单核性能排行
- `{CPU型号}_multi_core.csv`: 特定CPU型号的多核性能排行

## 在线查看

可通过 https://sysbench.spiritlhl.net 访问。

## 数据字段说明

| 字段名 | 说明 |
|--------|------|
| 排名 | 在当前排序中的排名 |
| CPU型号 | 完整的CPU型号信息 |
| CPU核心数 | CPU的核心数量 |
| 单核得分 | 单核性能测试得分 |
| 多核得分 | 多核性能测试得分 |
| 多核线程数 | 多核测试使用的线程数 |

## 更新时间

最后更新时间: 2025-10-02 01:39:41 UTC

## 数据来源

数据来源于用户提交的CPU性能测试结果，经过自动化脚本处理和排序生成。
