---
name: 档案局
description: "备份/同步/回滚/灾难恢复.双盘互备,原址删不删备份.调用scripts/"
---

# 档案局(已加载至工作记忆)

## 备份铁律

### 1. 新项目立项
C:\wuji-projects\{项目名}\ <- 日常开发+每次改动的.wuji-backups/
E:\wuji-projects\{项目名}\ <- robocopy /MIR镜像(调用wuji-e-sync.ps1)
原地址文件被删->备份文件夹不动(除非用户要求)

### 2. 每次改动前(调用wuji-backup.py)
备份原文件到.wuji-backups/{日期}_{时间}_{变更描述}/
更新CHANGELOG.md(描述/文件/原因/回滚指令)

### 3. 灾难恢复(调用wuji-restore.ps1)
系统重做->GitHub拉无极军团skill->从E:\wuji-projects\恢复所有项目

## 目录结构
C:\wuji-projects\{项目}\ <- .wuji-backups/{日期}_{描述}/ + CHANGELOG.md
E:\wuji-projects\{项目}\ <- 镜像, robocopy /MIR

## 回滚
档案局查CHANGELOG->找到目标版本->恢复->标记完成

## 自动激活口令
sop(archive.md) + 脚本参考
