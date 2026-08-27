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

def modify_go_mod(filepath):
    """
    Modify go.mod to remove privatespeedtest (and optional security) dependencies.
    Automatically matches module names regardless of version or indirect comment.
    """
    content = read_file(filepath)

    # Modules to remove
    remove_modules = [
        r'github\.com/oneclickvirt/privatespeedtest',
        r'github\.com/oneclickvirt/security',
    ]

    for mod in remove_modules:
        # Remove full require line (with or without // indirect)
        content = re.sub(
            rf'^[ \t]*{mod}[ \t]+v[^\s]+(?:[ \t]+// indirect)?[ \t]*\n',
            '',
            content,
            flags=re.MULTILINE
        )

    write_file(filepath, content)
    print(f"✓ Removed privatespeedtest/security from {filepath}")


def remove_vendor_tree(path='vendor'):
    """Remove the private-module vendor snapshot from the public branch."""
    if os.path.isdir(path):
        shutil.rmtree(path)
        print(f"✓ Removed {path}/ from public branch")


def remove_code_block(lines, start_marker, end_condition='empty_line'):
    """
    Remove code block from lines starting with start_marker.
    
    Args:
        lines: List of file lines
        start_marker: String or list of strings to identify block start
        end_condition: 'empty_line' (default) or 'closing_brace' or custom function
    
    Returns:
        Modified lines with the block removed
    """
    if isinstance(start_marker, str):
        start_marker = [start_marker]
    
    result = []
    skip_mode = False
    brace_depth = 0
    
    i = 0
    while i < len(lines):
        line = lines[i]
        
        # Check if we should start skipping
        if not skip_mode:
            for marker in start_marker:
                if marker in line:
                    skip_mode = True
                    if end_condition == 'closing_brace':
                        # Count opening braces on the function declaration line
                        brace_depth = line.count('{') - line.count('}')
                    break
            
            if not skip_mode:
                result.append(line)
        else:
            # We're in skip mode
            if end_condition == 'empty_line':
                # Skip until we find an empty line
                if line.strip() == '':
                    skip_mode = False
                    # Don't add the empty line, continue to next
            elif end_condition == 'closing_brace':
                # Track brace depth
                brace_depth += line.count('{') - line.count('}')
                if brace_depth == 0 and '}' in line:
                    # Function ended, skip until next empty line
                    end_condition = 'empty_line'
        
        i += 1
    
    return result


def modify_speed_go(filepath):
    """
    Replace the private-only speed implementation with the public implementation.

    The two files intentionally share the same API but use complementary build
    tags. Copying the already-public implementation keeps this transformation
    independent of comments or formatting in the private file.
    """
    public_filepath = os.path.join(os.path.dirname(filepath), 'speed_public.go')
    if not os.path.exists(public_filepath):
        raise FileNotFoundError(f"Public speed implementation not found: {public_filepath}")

    content = read_file(public_filepath)
    content, replacements = re.subn(
        r'(?m)^//go:build\s+ecs_public\s*$',
        '//go:build !ecs_public',
        content,
        count=1,
    )
    if replacements != 1:
        raise ValueError(f"Unexpected build tag in {public_filepath}")
    content = re.sub(
        r'(?m)^// \+build\s+ecs_public\s*$',
        '// +build !ecs_public',
        content,
        count=1,
    )
    forbidden = re.compile(
        r'privatespeedtest|privateSpeed|privatepst|private[_-]speed',
        flags=re.IGNORECASE,
    )
    if forbidden.search(content):
        raise ValueError(f"Public speed implementation contains restricted markers: {public_filepath}")

    write_file(filepath, content)
    print(f"✓ Replaced private speed implementation in {filepath}")

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
            r', binary files compiled in \[securityCheck\][^\)]*\)',
            ', but open sourced',
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
            r'二进制文件编译至 \[securityCheck\][^\)]*\)',
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


def sanitize_public_markdown(root='.'):
    """Remove restricted speed-test implementation details from public Markdown."""
    standalone = re.compile(
        r'^[^\r\n]*(?:privatespeedtest|private[ \t-]+carrier(?:[ \t-]+speed)?[ \t-]+nodes?|'
        r'私有(?:国内|三网)?测速(?:节点)?|私有节点|备用候选)[^\r\n]*(?:\r?\n|$)',
        flags=re.IGNORECASE | re.MULTILINE,
    )
    for directory, subdirectories, filenames in os.walk(root):
        subdirectories[:] = [
            name for name in subdirectories
            if name not in {'.git', 'vendor', '.cache', '.tmp'}
        ]
        for filename in filenames:
            if not filename.lower().endswith(('.md', '.mdx')):
                continue
            filepath = os.path.join(directory, filename)
            content = read_file(filepath)
            original = content

            # Preserve the surrounding public speed-test sentence while dropping
            # the restricted component name and its selection/registry details.
            content = re.sub(
                r'(?i)\s*Private[ \t-]+carrier(?:[ \t-]+speed)?[ \t-]+nodes?\s+from\s+`?privatespeedtest[^.\r\n]*\.\s*',
                ' ',
                content,
            )
            content = re.sub(
                r'，同时融合[^。\r\n]*(?:privatespeedtest|私有国内测速节点)[^。\r\n]*?；',
                '；',
                content,
                flags=re.IGNORECASE,
            )
            content = re.sub(
                r'(?i)\s*[（(](?:without private dependencies)[）)]',
                '',
                content,
            )
            content = re.sub(r'(?i)\bwithout private dependencies\b', '', content)
            content = re.sub(r'不含私有依赖', '', content)
            content = standalone.sub('', content)
            content = re.sub(r'\n{3,}', '\n\n', content)

            if content != original:
                write_file(filepath, content)
                print(f"✓ Sanitized public documentation: {filepath}")


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
    sanitize_public_markdown()
    print()
    
    # Modify go.mod
    print("Modifying go.mod...")
    modify_go_mod('go.mod')
    remove_vendor_tree()
    print()
        
    print("✓ All modifications completed successfully!")
    print()
    print("Next steps:")
    print("1. Run 'go mod tidy' to clean up dependencies")
    print("2. Run 'go build -o maintest' to verify compilation")
    print("3. Test the binary with: ./maintest -menu=false -l en -security=false -upload=false")


if __name__ == '__main__':
    main()
