# target-range.py - 无极军团打靶场系统

import os, sys, json, re, glob

class TargetRange:
    """打靶场 - 沙盒测试环境"""

    def __init__(self):
        self.checks_passed = 0
        self.checks_total = 0
        self.checks_failed = []
        self.checks_warning = []
        self.report_lines = []

    def check(self, name, result, detail=""):
        """执行一项检查"""
        self.checks_total += 1
        if result:
            self.checks_passed += 1
            self.report_lines.append("  [PASS] %s" % name)
        elif result is None:
            self.checks_warning.append(name)
            self.report_lines.append("  [WARN] %s - %s" % (name, detail))
        else:
            self.checks_failed.append(name)
            self.report_lines.append("  [FAIL] %s - %s" % (name, detail))

    def scan_code(self, filepath):
        """代码安全扫描 - 检查常见风险模式"""
        if not os.path.exists(filepath):
            self.check("文件存在性", False, "文件不存在")
            return

        with open(filepath, "r", encoding="utf-8", errors="replace") as f:
            content = f.read()

        self.check("无命令注入",
            not any(kw in content for kw in ["os.system(", "subprocess.call(", "subprocess.Popen(",
                                              "eval(", "exec(", "__import__(", "compile("]),
            "包含动态代码执行")

        self.check("文件操作安全",
            "shutil.rmtree" not in content and "os.remove(" not in content,
            "包含删除文件操作（需确认）")

        key_patterns = [
            r'(?:api[_-]?key|apikey|secret|token|password)[\'"]\s*[:=]\s*[\'"](?!\s*$|os\.environ|os\.getenv)',
            r'(?:AKIA[0-9A-Z]{16}|sk-[a-zA-Z0-9]{32,}|ghp_[a-zA-Z0-9]{36})'
        ]
        has_hardcoded_key = False
        for pat in key_patterns:
            if re.search(pat, content, re.IGNORECASE):
                has_hardcoded_key = True
                break
        self.check("无硬编码密钥", not has_hardcoded_key, "发现可能的硬编码密钥")

        size = os.path.getsize(filepath)
        self.check("文件大小合理", size < 5 * 1024 * 1024, "文件超过5MB（%.1fMB）" % (size / 1024 / 1024))

        if content.startswith("#") or content.startswith("/*"):
            self.check("代码头部检查", True)

        dangerous_imports = ["requests", "urllib", "ftplib", "telnetlib", "smtplib"]
        found_dangerous = [m for m in dangerous_imports if "import %s" % m in content or "from %s" % m in content]
        self.check("网络模块审查", len(found_dangerous) == 0,
                    "导入了网络模块: %s（需确认）" % ", ".join(found_dangerous))

    def scan_plugin(self, plugin_dir):
        """ComfyUI 插件扫描"""
        if not os.path.isdir(plugin_dir):
            self.check("插件目录存在", False, "目录不存在")
            return

        files = os.listdir(plugin_dir)
        self.check("插件目录非空", len(files) > 0)

        has_init = "__init__.py" in files
        has_main = any(f.endswith(".py") for f in files)
        self.check("插件入口文件", has_init or has_main, "缺少 __init__.py 或 .py 文件")

        py_files = glob.glob(os.path.join(plugin_dir, "**", "*.py"), recursive=True)
        for pyf in py_files[:5]:  # 最多扫 5 个文件
            self.scan_code(pyf)

        req_file = os.path.join(plugin_dir, "requirements.txt")
        if os.path.exists(req_file):
            with open(req_file, "r") as f:
                deps = [l.strip() for l in f if l.strip() and not l.startswith("#")]
            self.check("依赖项检查", len(deps) <= 20, "依赖过多（%d项）" % len(deps))

    def scan_dependency(self, dep_name):
        """依赖安全检查"""
        dangerous_keywords = ["eval", "exec", "compile", "base64", "obfuscat", "crypt"]
        is_dangerous = any(kw in dep_name.lower() for kw in dangerous_keywords)
        self.check("依赖名安全", not is_dangerous, "依赖名包含可疑关键字")
        self.check("依赖名长度合理", len(dep_name) < 100, "依赖名过长")

    def scan_config(self, config_path):
        """配置文件检查"""
        if not os.path.exists(config_path):
            self.check("配置文件存在", False, "文件不存在")
            return
        with open(config_path, "r", encoding="utf-8", errors="replace") as f:
            content = f.read()
        self.check("配置文件可读", len(content) > 0)
        self.check("无敏感配置",
            "password" not in content.lower() and "secret" not in content.lower() and "token" not in content.lower(),
            "配置中包含敏感关键词，请确认是否需要加密")

    def scan_permission(self, operation_desc):
        """权限等级评估"""
        high_risk_keywords = ["delete", "remove", "rm", "format", "force push", "reset", "drop table",
                              "环境变量", "系统设置", "注册表"]
        medium_risk_keywords = ["install", "git push", "modify config", "覆盖", "重写"]

        desc_lower = operation_desc.lower()
        if any(kw in desc_lower for kw in high_risk_keywords):
            level = "[!R!] 高危"
            self.check("权限等级评估", False, level + " - 需要用户确认")
        elif any(kw in desc_lower for kw in medium_risk_keywords):
            level = "[!O!] 中等"
            self.check("权限等级评估", None, level + " - 建议确认")
        else:
            level = "[!G!] 无害"
            self.check("权限等级评估", True, level)

    def generate_report(self):
        """生成打靶场报告"""
        passed = self.checks_passed
        total = self.checks_total
        failed = len(self.checks_failed)
        warned = len(self.checks_warning)
        score = int(passed / total * 100) if total > 0 else 0

        print("=" * 50)
        print("  >>  无极军团 - 打靶场测试报告")
        print("=" * 50)
        print("")
        for line in self.report_lines:
            print(line)
        print("")
        print("-" * 50)
        print("  总计: %d 项 | 通过: %d | 警告: %d | 失败: %d" % (total, passed, warned, failed))
        print("  评分: %d/100" % score)

        if score >= 80:
            verdict = "[OK] 通过 - 可以执行"
        elif score >= 60:
            verdict = "[--] 有条件通过 - 处理警告后执行"
        else:
            verdict = "[XX] 未通过 - 必须修复失败项"
        print("  结论: %s" % verdict)
        print("=" * 50)
        return {
            "score": score,
            "passed": passed,
            "total": total,
            "failed": failed,
            "warnings": warned,
            "verdict": verdict,
            "details": self.report_lines
        }

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("无极军团打靶场系统 v1.0")
        print("")
        print("用法:")
        print("  python target-range.py code <file>           # 扫描代码文件")
        print("  python target-range.py plugin <dir>          # 扫描ComfyUI插件")
        print("  python target-range.py dependency <name>     # 检查依赖")
        print("  python target-range.py config <file>         # 检查配置")
        print("  python target-range.py permission <操作描述>  # 评估权限等级")
        print("  python target-range.py skill <skill_dir>     # 扫描skill目录")
        sys.exit(1)

    target_type = sys.argv[1]
    target_path = sys.argv[2]

    tr = TargetRange()

    if target_type == "code":
        tr.scan_code(target_path)
    elif target_type == "plugin":
        tr.scan_plugin(target_path)
    elif target_type == "dependency":
        tr.scan_dependency(target_path)
    elif target_type == "config":
        tr.scan_config(target_path)
    elif target_type == "permission":
        tr.scan_permission(target_path)
    elif target_type == "skill":
        tr.scan_code(target_path)  # 复用代码扫描
    else:
        print("[ERROR] Unknown target type: %s" % target_type)
        sys.exit(1)

    tr.generate_report()
