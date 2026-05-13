# wuji-backup.py — 无极军团智能备份系统

import os, sys, shutil, glob
from datetime import datetime

MAX_BACKUPS = 10  # 最多保留 10 份备份

def get_backup_dir(filepath):
    """备份文件放在原文件同级目录下的 .wuji-backups/ 文件夹"""
    basedir = os.path.dirname(os.path.abspath(filepath)) or "."
    backup_dir = os.path.join(basedir, ".wuji-backups")
    os.makedirs(backup_dir, exist_ok=True)
    return backup_dir

def create_backup(filepath):
    """为指定文件创建备份"""
    if not os.path.exists(filepath):
        print("[SKIP] File does not exist: %s" % filepath)
        return None

    basename = os.path.basename(filepath)
    now = datetime.now()
    timestamp = now.strftime("%m%d%H%M%S")
    backup_name = '%s_%s_%03d.bak' % (basename, timestamp, now.microsecond // 1000)
    backup_dir = get_backup_dir(filepath)
    backup_path = os.path.join(backup_dir, backup_name)

    shutil.copy2(filepath, backup_path)
    print("[OK] Backup created: %s" % backup_path)

    rotate_backups(filepath)

    return backup_path

def rotate_backups(filepath):
    """保留最近的 MAX_BACKUPS 份备份，删除更旧的"""
    basename = os.path.basename(filepath)
    backup_dir = get_backup_dir(filepath)
    pattern = os.path.join(backup_dir, "%s_*.bak" % basename)

    backups = sorted(glob.glob(pattern), key=os.path.getmtime)

    if len(backups) > MAX_BACKUPS:
        to_delete = backups[:len(backups) - MAX_BACKUPS]
        for old in to_delete:
            os.remove(old)
            print("[CLEAN] Removed old backup: %s" % os.path.basename(old))
        print("[INFO] Kept %d latest backups for %s" % (MAX_BACKUPS, basename))

def list_backups(filepath):
    """列出指定文件的所有备份"""
    basename = os.path.basename(filepath)
    backup_dir = get_backup_dir(filepath)
    pattern = os.path.join(backup_dir, "%s_*.bak" % basename)

    backups = sorted(glob.glob(pattern), key=os.path.getmtime, reverse=True)
    if not backups:
        print("[INFO] No backups found for: %s" % basename)
        return []

    print("[INFO] Backups for %s (%d total, max %d):" % (basename, len(backups), MAX_BACKUPS))
    for i, b in enumerate(backups, 1):
        mtime = datetime.fromtimestamp(os.path.getmtime(b))
        size = os.path.getsize(b)
        print("  %d. %s (%s, %d bytes)" % (i, os.path.basename(b), mtime.strftime("%m-%d %H:%M"), size))
    return backups

def restore_backup(filepath, version=None):
    """从备份恢复文件
    version: None=最新, 1=第1新的, 2=第2新的...
    """
    backups = list_backups(filepath)
    if not backups:
        print("[ERROR] No backups to restore from")
        return False

    if version is None:
        restore_from = backups[0]  # 最新的
    elif 1 <= version <= len(backups):
        restore_from = backups[version - 1]
    else:
        print("[ERROR] Version %d not found (available: 1-%d)" % (version, len(backups)))
        return False

    if os.path.exists(filepath):
        create_backup(filepath)

    shutil.copy2(restore_from, filepath)
    print("[OK] Restored: %s <- %s" % (filepath, os.path.basename(restore_from)))
    return True

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Wuji Backup System v1.0")
        print("Usage:")
        print("  python wuji-backup.py backup <file>      # Create backup (default)")
        print("  python wuji-backup.py list <file>        # List backups")
        print("  python wuji-backup.py restore <file> [n] # Restore (n=version, default latest)")
        print("  python wuji-backup.py clean <file>       # Force cleanup old backups")
        sys.exit(1)

    cmd = sys.argv[1]
    if cmd == "backup" and len(sys.argv) > 2:
        create_backup(sys.argv[2])
    elif cmd == "list" and len(sys.argv) > 2:
        list_backups(sys.argv[2])
    elif cmd == "restore" and len(sys.argv) > 2:
        version = int(sys.argv[3]) if len(sys.argv) > 3 else None
        restore_backup(sys.argv[2], version)
    elif cmd == "clean" and len(sys.argv) > 2:
        rotate_backups(sys.argv[2])
    else:
        create_backup(sys.argv[1])
