#!/usr/bin/env python3
"""
Version utilities for automated release workflow.

This module provides utility functions for version management including:
- Semantic version validation
- Git tag existence checking
- Version comparison and validation
"""

import re
import subprocess
import sys
from typing import Optional, Tuple


def validate_semantic_version(version: str) -> bool:
    """
    Validate if a version string follows semantic versioning format.
    
    Args:
        version: Version string to validate (e.g., "1.2.3", "0.1.0")
        
    Returns:
        bool: True if version is valid semantic version, False otherwise
    """
    # Semantic version pattern: MAJOR.MINOR.PATCH with optional pre-release and build metadata
    # For this project, we'll use strict format: X.Y.Z where X, Y, Z are non-negative integers
    pattern = r'^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$'
    return bool(re.match(pattern, version.strip()))


def read_version_from_file(file_path: str = "VERSION") -> Optional[str]:
    """
    Read version from VERSION file and validate format.
    
    Args:
        file_path: Path to VERSION file (default: "VERSION")
        
    Returns:
        str: Version string if valid, None if file doesn't exist or invalid format
    """
    try:
        with open(file_path, 'r') as f:
            version = f.read().strip()
            
        if validate_semantic_version(version):
            return version
        else:
            print(f"Error: Invalid semantic version format in {file_path}: {version}")
            return None
            
    except FileNotFoundError:
        print(f"Error: {file_path} file not found")
        return None
    except Exception as e:
        print(f"Error reading {file_path}: {e}")
        return None


def git_tag_exists(version: str) -> bool:
    """
    Check if a git tag already exists for the given version.
    
    Args:
        version: Version string (will be prefixed with 'v' for tag check)
        
    Returns:
        bool: True if tag exists, False otherwise
    """
    tag_name = f"v{version}"
    
    try:
        # Check if tag exists locally
        result = subprocess.run(
            ["git", "tag", "-l", tag_name],
            capture_output=True,
            text=True,
            check=False
        )
        
        if result.returncode == 0 and result.stdout.strip():
            return True
            
        # Also check remote tags to be thorough
        result = subprocess.run(
            ["git", "ls-remote", "--tags", "origin", tag_name],
            capture_output=True,
            text=True,
            check=False
        )
        
        return result.returncode == 0 and result.stdout.strip() != ""
        
    except Exception as e:
        print(f"Error checking git tag existence: {e}")
        return False


def compare_versions(version1: str, version2: str) -> int:
    """
    Compare two semantic versions.
    
    Args:
        version1: First version string
        version2: Second version string
        
    Returns:
        int: -1 if version1 < version2, 0 if equal, 1 if version1 > version2
    """
    if not validate_semantic_version(version1) or not validate_semantic_version(version2):
        raise ValueError("Both versions must be valid semantic versions")
    
    def parse_version(version: str) -> Tuple[int, int, int]:
        parts = version.split('.')
        return (int(parts[0]), int(parts[1]), int(parts[2]))
    
    v1_parts = parse_version(version1)
    v2_parts = parse_version(version2)
    
    if v1_parts < v2_parts:
        return -1
    elif v1_parts > v2_parts:
        return 1
    else:
        return 0


def get_git_commit_version_change(commit_hash: str = "HEAD") -> Optional[str]:
    """
    Check if VERSION file changed in a specific commit and return the new version.
    
    Args:
        commit_hash: Git commit hash to check (default: "HEAD")
        
    Returns:
        str: New version if VERSION file changed, None otherwise
    """
    try:
        # Check if VERSION file was modified in the commit
        result = subprocess.run(
            ["git", "diff", "--name-only", f"{commit_hash}^", commit_hash],
            capture_output=True,
            text=True,
            check=True
        )
        
        changed_files = result.stdout.strip().split('\n')
        
        if "VERSION" in changed_files:
            # Get the new version from the commit
            result = subprocess.run(
                ["git", "show", f"{commit_hash}:VERSION"],
                capture_output=True,
                text=True,
                check=True
            )
            
            version = result.stdout.strip()
            if validate_semantic_version(version):
                return version
            else:
                print(f"Error: Invalid version format in commit {commit_hash}: {version}")
                return None
        
        return None
        
    except subprocess.CalledProcessError as e:
        print(f"Error checking git commit changes: {e}")
        return None
    except Exception as e:
        print(f"Unexpected error: {e}")
        return None


def main():
    """
    Command-line interface for version utilities.
    
    Usage examples:
        python version_utils.py validate 1.2.3
        python version_utils.py check-tag 1.2.3
        python version_utils.py read-version
        python version_utils.py compare 1.2.3 1.3.0
    """
    if len(sys.argv) < 2:
        print("Usage: python version_utils.py <command> [args...]")
        print("Commands:")
        print("  validate <version>     - Validate semantic version format")
        print("  check-tag <version>    - Check if git tag exists for version")
        print("  read-version [file]    - Read and validate version from file")
        print("  compare <v1> <v2>      - Compare two versions")
        print("  check-commit [hash]    - Check if VERSION changed in commit")
        sys.exit(1)
    
    command = sys.argv[1]
    
    if command == "validate":
        if len(sys.argv) != 3:
            print("Usage: python version_utils.py validate <version>")
            sys.exit(1)
        
        version = sys.argv[2]
        if validate_semantic_version(version):
            print(f"✓ Version {version} is valid")
            sys.exit(0)
        else:
            print(f"✗ Version {version} is invalid")
            sys.exit(1)
    
    elif command == "check-tag":
        if len(sys.argv) != 3:
            print("Usage: python version_utils.py check-tag <version>")
            sys.exit(1)
        
        version = sys.argv[2]
        if git_tag_exists(version):
            print(f"✗ Tag v{version} already exists")
            sys.exit(1)
        else:
            print(f"✓ Tag v{version} does not exist")
            sys.exit(0)
    
    elif command == "read-version":
        file_path = sys.argv[2] if len(sys.argv) > 2 else "VERSION"
        version = read_version_from_file(file_path)
        if version:
            print(version)
            sys.exit(0)
        else:
            sys.exit(1)
    
    elif command == "compare":
        if len(sys.argv) != 4:
            print("Usage: python version_utils.py compare <version1> <version2>")
            sys.exit(1)
        
        v1, v2 = sys.argv[2], sys.argv[3]
        try:
            result = compare_versions(v1, v2)
            if result == -1:
                print(f"{v1} < {v2}")
            elif result == 0:
                print(f"{v1} = {v2}")
            else:
                print(f"{v1} > {v2}")
            sys.exit(0)
        except ValueError as e:
            print(f"Error: {e}")
            sys.exit(1)
    
    elif command == "check-commit":
        commit_hash = sys.argv[2] if len(sys.argv) > 2 else "HEAD"
        version = get_git_commit_version_change(commit_hash)
        if version:
            print(version)
            sys.exit(0)
        else:
            print("No version change detected")
            sys.exit(1)
    
    else:
        print(f"Unknown command: {command}")
        sys.exit(1)


if __name__ == "__main__":
    main()