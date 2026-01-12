#!/usr/bin/env python3
"""
Script to create public branch by removing security dependencies and references.
This script properly handles Go file modifications to ensure the code can compile.
"""

import re
import os
import sys
import shutil


def read_file(filepath):
    """Read file content."""
    with open(filepath, 'r', encoding='utf-8') as f:
        return f.read()


def write_file(filepath, content):
    """Write content to file."""
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)


def modify_speed_go(filepath):
    """
    Brutally remove privatespeedtest-related code from speed.go
    by deleting known blocks from upstream source.
    """
    content = read_file(filepath)

    content = re.sub(
        r'\n\s*"github\.com/oneclickvirt/privatespeedtest/pst"\s*\n',
        '\n',
        content,
        flags=re.MULTILINE
    )

    content = re.sub(
        r'// formatString 格式化字符串到指定宽度[\s\S]*?return nil\n}\n',
        '',
        content,
        flags=re.MULTILINE
    )

    content = re.sub(
        r'^[ \t]*// 对于三网测速（cmcc、cu、ct），优先使用 privatespeedtest 进行私有测速[\s\S]*?\n\s*\n',
        '\n',
        content,
        flags=re.MULTILINE
    )

    write_file(filepath, content)
    print(f"✓ Cleanly removed privatespeedtest from {filepath}")

def modify_utils_go(filepath):
    """
    Modify utils/utils.go to:
    1. Replace security/network import with basics/network
    2. Replace SecurityUploadToken usage with hardcoded token
    """
    content = read_file(filepath)
    
    # Replace import
    content = re.sub(
        r'"github\.com/oneclickvirt/security/network"',
        r'"github.com/oneclickvirt/basics/network"',
        content
    )
    
    # Replace token usage - find the exact line and replace it
    content = re.sub(
        r'\ttoken := network\.SecurityUploadToken',
        r'\ttoken := "OvwKx5qgJtf7PZgCKbtyojSU.MTcwMTUxNzY1MTgwMw"',
        content
    )
    
    # Update title for public version
    content = re.sub(
        r'VPS融合怪测试',
        r'VPS融合怪测试(非官方编译)',
        content
    )
    content = re.sub(
        r'VPS Fusion Monster Test',
        r'VPS Fusion Monster Test (Unofficial)',
        content
    )
    
    write_file(filepath, content)
    print(f"✓ Modified {filepath}")


def modify_params_go(filepath):
    """
    Modify internal/params/params.go to change security flag default to false.
    """
    content = read_file(filepath)
    
    # Change default value in struct initialization
    content = re.sub(
        r'(\s+SecurityTestStatus:\s+)true,',
        r'\1false,',
        content
    )
    
    # Change flag default value
    content = re.sub(
        r'(c\.GoecsFlag\.BoolVar\(&c\.SecurityTestStatus, "security", )true(, "Enable/Disable security test"\))',
        r'\1false\2',
        content
    )
    
    write_file(filepath, content)
    print(f"✓ Modified {filepath}")


# def modify_go_mod(filepath):
#     """
#     Modify go.mod to remove security and privatespeedtest dependencies.
#     """
#     content = read_file(filepath)
    
#     # Remove security dependency from require section
#     content = re.sub(
#         r'\s+github\.com/oneclickvirt/security v[^\n]+\n',
#         '',
#         content
#     )
    
#     # Remove privatespeedtest dependency from require section (including indirect)
#     content = re.sub(
#         r'\s+github\.com/oneclickvirt/privatespeedtest v[^\n]+\n',
#         '',
#         content
#     )
    
#     write_file(filepath, content)
#     print(f"✓ Modified {filepath}")


def modify_readme(filepath, is_english=False):
    """
    Modify README files to update Go version and security status.
    """
    content = read_file(filepath)
    
    # Extract Go version from go.mod
    go_mod_content = read_file('go.mod')
    go_version_match = re.search(r'^go (\d+\.\d+(?:\.\d+)?)', go_mod_content, re.MULTILINE)
    
    if not go_version_match:
        print(f"⚠ Warning: Could not extract Go version from go.mod")
        return
    
    go_version = go_version_match.group(1)
    
    if is_english:
        # Update Go version in English README
        content = re.sub(
            r'Select go \d+\.\d+\.\d+ version to install',
            f'Select go {go_version} version to install',
            content
        )
        
        # Update security status
        content = re.sub(
            r'but binary files compiled in \[securityCheck\][^\)]*\)',
            'but open sourced',
            content
        )
        
        # Update help text for security flag
        content = re.sub(
            r'security\s+Enable/Disable security test \(default true\)',
            'security        Enable/Disable security test (default false)',
            content
        )
    else:
        # Update Go version in Chinese README
        content = re.sub(
            r'选择 go \d+\.\d+\.\d+ 的版本进行安装',
            f'选择 go {go_version} 的版本进行安装',
            content
        )
        
        # Update security status
        content = re.sub(
            r'但二进制文件编译至 \[securityCheck\][^\)]*\)',
            '但已开源',
            content
        )
        
        # Update help text for security flag
        content = re.sub(
            r'security\s+Enable/Disable security test \(default true\)',
            'security        Enable/Disable security test (default false)',
            content
        )
    
    write_file(filepath, content)
    print(f"✓ Modified {filepath}")


def main():
    """Main function to process all files."""
    print("Starting public branch creation process...")
    print()
    
    # Check if we're in the right directory
    if not os.path.exists('go.mod'):
        print("Error: go.mod not found. Please run this script from the project root.")
        sys.exit(1)
    
    # Modify Go source files
    print("Modifying Go source files...")
    modify_speed_go('internal/tests/speed.go')
    modify_utils_go('utils/utils.go')
    modify_params_go('internal/params/params.go')
    print()
    
    # Modify go.mod
    # print("Modifying go.mod...")
    # modify_go_mod('go.mod')
    # print()
    
    # Modify README files
    print("Modifying README files...")
    modify_readme('README.md', is_english=False)
    modify_readme('README_EN.md', is_english=True)
    print()
    
    print("✓ All modifications completed successfully!")
    print()
    print("Next steps:")
    print("1. Run 'go mod tidy' to clean up dependencies")
    print("2. Run 'go build -o maintest' to verify compilation")
    print("3. Test the binary with: ./maintest -menu=false -l en -security=false -upload=false")


if __name__ == '__main__':
    main()
