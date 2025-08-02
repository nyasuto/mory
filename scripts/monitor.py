#!/usr/bin/env python3
"""
Monitoring and health check script for Mory Server
Provides system health monitoring and alerting capabilities
"""

import argparse
import json
import sqlite3
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

import httpx


class MoryMonitor:
    """Health monitoring for Mory Server"""

    def __init__(
        self, base_url: str = "http://localhost:8080", data_dir: str = "/opt/mory-server/data"
    ):
        """Initialize monitor"""
        self.base_url = base_url.rstrip("/")
        self.data_dir = Path(data_dir)
        self.client = httpx.Client(timeout=30.0)

    def check_api_health(self) -> dict:
        """Check API endpoint health"""
        health_info = {"status": "unknown", "response_time": None, "error": None}

        try:
            start_time = time.time()
            response = self.client.get(f"{self.base_url}/api/health")
            response_time = time.time() - start_time

            health_info["response_time"] = round(response_time * 1000, 2)  # ms

            if response.status_code == 200:
                health_info["status"] = "healthy"
                health_info["data"] = response.json()
            else:
                health_info["status"] = "unhealthy"
                health_info["error"] = f"HTTP {response.status_code}"

        except httpx.ConnectError:
            health_info["status"] = "unreachable"
            health_info["error"] = "Connection failed"
        except httpx.TimeoutException:
            health_info["status"] = "timeout"
            health_info["error"] = "Request timeout"
        except Exception as e:
            health_info["status"] = "error"
            health_info["error"] = str(e)

        return health_info

    def check_detailed_health(self) -> dict:
        """Check detailed API health"""
        health_info = {"status": "unknown", "response_time": None, "error": None}

        try:
            start_time = time.time()
            response = self.client.get(f"{self.base_url}/api/health/detailed")
            response_time = time.time() - start_time

            health_info["response_time"] = round(response_time * 1000, 2)  # ms

            if response.status_code == 200:
                health_info["status"] = "healthy"
                health_info["data"] = response.json()
            else:
                health_info["status"] = "unhealthy"
                health_info["error"] = f"HTTP {response.status_code}"

        except Exception as e:
            health_info["status"] = "error"
            health_info["error"] = str(e)

        return health_info

    def check_database_health(self) -> dict:
        """Check database health directly"""
        db_info = {"status": "unknown", "size": None, "record_count": None, "error": None}

        db_path = self.data_dir / "mory.db"

        try:
            if not db_path.exists():
                db_info["status"] = "missing"
                db_info["error"] = "Database file not found"
                return db_info

            # Check file size
            db_info["size"] = db_path.stat().st_size

            # Check database integrity and get stats
            conn = sqlite3.connect(db_path)
            cursor = conn.cursor()

            # Integrity check
            cursor.execute("PRAGMA integrity_check")
            integrity_result = cursor.fetchone()[0]

            if integrity_result != "ok":
                db_info["status"] = "corrupted"
                db_info["error"] = f"Integrity check failed: {integrity_result}"
                return db_info

            # Get record count
            cursor.execute("SELECT COUNT(*) FROM memories")
            db_info["record_count"] = cursor.fetchone()[0]

            # Get last update time
            cursor.execute("SELECT MAX(updated_at) FROM memories")
            last_update = cursor.fetchone()[0]
            if last_update:
                db_info["last_update"] = last_update

            conn.close()
            db_info["status"] = "healthy"

        except sqlite3.Error as e:
            db_info["status"] = "error"
            db_info["error"] = f"Database error: {e}"
        except Exception as e:
            db_info["status"] = "error"
            db_info["error"] = str(e)

        return db_info

    def check_service_status(self) -> dict:
        """Check systemd service status"""
        service_info = {
            "status": "unknown",
            "active": None,
            "enabled": None,
            "pid": None,
            "memory_usage": None,
            "cpu_usage": None,
            "error": None,
        }

        try:
            # Check service status
            result = subprocess.run(
                ["systemctl", "is-active", "mory-server"], capture_output=True, text=True
            )
            service_info["active"] = result.stdout.strip() == "active"

            # Check if enabled
            result = subprocess.run(
                ["systemctl", "is-enabled", "mory-server"], capture_output=True, text=True
            )
            service_info["enabled"] = result.stdout.strip() == "enabled"

            # Get detailed status
            result = subprocess.run(
                ["systemctl", "show", "mory-server", "--property=MainPID,MemoryCurrent"],
                capture_output=True,
                text=True,
            )

            for line in result.stdout.split("\n"):
                if line.startswith("MainPID="):
                    pid = line.split("=")[1]
                    service_info["pid"] = int(pid) if pid != "0" else None
                elif line.startswith("MemoryCurrent="):
                    memory = line.split("=")[1]
                    if memory != "[not set]":
                        service_info["memory_usage"] = int(memory)

            service_info["status"] = "healthy" if service_info["active"] else "inactive"

        except subprocess.CalledProcessError as e:
            service_info["status"] = "error"
            service_info["error"] = f"Command failed: {e}"
        except Exception as e:
            service_info["status"] = "error"
            service_info["error"] = str(e)

        return service_info

    def check_disk_space(self) -> dict:
        """Check disk space in data directory"""
        disk_info = {
            "status": "unknown",
            "total": None,
            "used": None,
            "free": None,
            "usage_percent": None,
            "error": None,
        }

        try:
            result = subprocess.run(["df", str(self.data_dir)], capture_output=True, text=True)

            lines = result.stdout.strip().split("\n")
            if len(lines) >= 2:
                fields = lines[1].split()
                if len(fields) >= 6:
                    disk_info["total"] = int(fields[1]) * 1024  # Convert to bytes
                    disk_info["used"] = int(fields[2]) * 1024
                    disk_info["free"] = int(fields[3]) * 1024
                    disk_info["usage_percent"] = int(fields[4].rstrip("%"))

                    # Determine status based on usage
                    if disk_info["usage_percent"] >= 90:
                        disk_info["status"] = "critical"
                    elif disk_info["usage_percent"] >= 80:
                        disk_info["status"] = "warning"
                    else:
                        disk_info["status"] = "healthy"

        except Exception as e:
            disk_info["status"] = "error"
            disk_info["error"] = str(e)

        return disk_info

    def run_full_health_check(self) -> dict:
        """Run comprehensive health check"""
        print("🔍 Running Mory Server health check...")

        health_report = {
            "timestamp": datetime.now().isoformat(),
            "overall_status": "unknown",
            "checks": {},
        }

        # Run all checks
        checks = {
            "api": self.check_api_health,
            "detailed_api": self.check_detailed_health,
            "database": self.check_database_health,
            "service": self.check_service_status,
            "disk": self.check_disk_space,
        }

        failed_checks = []

        for check_name, check_func in checks.items():
            print(f"  Checking {check_name}...")
            health_report["checks"][check_name] = check_func()

            status = health_report["checks"][check_name]["status"]
            if status not in ["healthy", "warning"]:
                failed_checks.append(check_name)

        # Determine overall status
        if not failed_checks:
            health_report["overall_status"] = "healthy"
        elif len(failed_checks) == len(checks):
            health_report["overall_status"] = "critical"
        else:
            health_report["overall_status"] = "degraded"

        return health_report

    def print_health_report(self, report: dict):
        """Print formatted health report"""
        print("\n" + "=" * 60)
        print("📊 MORY SERVER HEALTH REPORT")
        print("=" * 60)

        # Overall status
        status_emoji = {
            "healthy": "✅",
            "warning": "⚠️",
            "degraded": "🟡",
            "critical": "❌",
            "unknown": "❓",
        }

        print(
            f"Overall Status: {status_emoji.get(report['overall_status'], '❓')} {report['overall_status'].upper()}"
        )
        print(f"Timestamp: {report['timestamp']}")
        print()

        # Individual checks
        for check_name, check_result in report["checks"].items():
            status = check_result["status"]
            emoji = status_emoji.get(status, "❓")
            print(f"{emoji} {check_name.upper()}: {status}")

            if check_result.get("error"):
                print(f"   Error: {check_result['error']}")

            if check_name == "api" and check_result.get("response_time"):
                print(f"   Response time: {check_result['response_time']}ms")

            if check_name == "database":
                if check_result.get("record_count") is not None:
                    print(f"   Records: {check_result['record_count']:,}")
                if check_result.get("size"):
                    print(f"   Size: {self._format_size(check_result['size'])}")

            if check_name == "service":
                if check_result.get("pid"):
                    print(f"   PID: {check_result['pid']}")
                if check_result.get("memory_usage"):
                    print(f"   Memory: {self._format_size(check_result['memory_usage'])}")

            if check_name == "disk":
                if check_result.get("usage_percent") is not None:
                    print(f"   Usage: {check_result['usage_percent']}%")
                if check_result.get("free"):
                    print(f"   Free: {self._format_size(check_result['free'])}")

            print()

    def _format_size(self, size: int) -> str:
        """Format file size in human readable format"""
        for unit in ["B", "KB", "MB", "GB"]:
            if size < 1024.0:
                return f"{size:.1f} {unit}"
            size /= 1024.0
        return f"{size:.1f} TB"

    def close(self):
        """Close HTTP client"""
        self.client.close()


def main():
    """Main monitoring function"""
    parser = argparse.ArgumentParser(description="Mory Server Monitoring")
    parser.add_argument("--url", default="http://localhost:8080", help="Base URL for Mory server")
    parser.add_argument(
        "--data-dir", default="/opt/mory-server/data", help="Path to Mory data directory"
    )
    parser.add_argument("--json", action="store_true", help="Output results as JSON")
    parser.add_argument(
        "--check",
        choices=["api", "database", "service", "disk", "all"],
        default="all",
        help="Specific check to run",
    )

    args = parser.parse_args()

    monitor = MoryMonitor(args.url, args.data_dir)

    try:
        if args.check == "all":
            report = monitor.run_full_health_check()
        elif args.check == "api":
            report = {"checks": {"api": monitor.check_api_health()}}
        elif args.check == "database":
            report = {"checks": {"database": monitor.check_database_health()}}
        elif args.check == "service":
            report = {"checks": {"service": monitor.check_service_status()}}
        elif args.check == "disk":
            report = {"checks": {"disk": monitor.check_disk_space()}}

        if args.json:
            print(json.dumps(report, indent=2, default=str))
        else:
            monitor.print_health_report(report)

        # Exit with error code if not healthy
        if args.check == "all":
            return 0 if report["overall_status"] in ["healthy", "warning"] else 1
        else:
            check_result = list(report["checks"].values())[0]
            return 0 if check_result["status"] in ["healthy", "warning"] else 1

    except Exception as e:
        print(f"❌ Monitoring failed: {e}")
        return 1

    finally:
        monitor.close()


if __name__ == "__main__":
    sys.exit(main())
