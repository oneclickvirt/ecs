#!/usr/bin/env python3
"""Create a public branch without loading private repository modules."""

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
    # The public adapter intentionally keeps the PrivateSpeedPreloads type and
    # its no-op functions so the runner has the same API under both build tags.
    # Validate the actual dependency boundary instead of rejecting those
    # compatibility identifiers by substring.
    private_import = re.compile(
        r'github\.com/oneclickvirt/(?:privatespeedtest|security)(?:/|["`])',
        flags=re.IGNORECASE,
    )
    if private_import.search(content):
        raise ValueError(f"Public speed implementation imports a restricted module: {public_filepath}")

    write_file(filepath, content)
    print(f"✓ Replaced private speed implementation in {filepath}")


def activate_public_component(filepath):
    """Make a public-only implementation available to ordinary Go builds."""
    content = read_file(filepath)
    content, replacements = re.subn(
        r'(?m)^//go:build\s+ecs_public\s*\n(?:^// \+build\s+ecs_public\s*\n)?\n?',
        '',
        content,
        count=1,
    )
    if replacements != 1:
        raise ValueError(f"Unexpected public build tag in {filepath}")
    write_file(filepath, content)
    print(f"✓ Activated public implementation for default builds: {filepath}")


def remove_private_go_sources():
    """Remove source files whose dependencies are intentionally private."""
    private_files = [
        'api/components_security_private.go',
        'api/components_speed_private.go',
        'api/components_local_test.go',
        'goecs_private_test.go',
        'internal/tests/speed_private_registry.go',
        'internal/tests/speed_test.go',
    ]
    for filepath in private_files:
        if not os.path.isfile(filepath):
            raise FileNotFoundError(f"Private source expected by public generator is missing: {filepath}")
        os.remove(filepath)
        print(f"✓ Removed private source: {filepath}")


def remove_private_delivery_artifacts():
    """Remove workflows that would load private modules in the public branch."""
    private_artifacts = [
        '.back/create_public_branch.py',
        '.github/workflows/build_binary.yaml',
        '.github/workflows/build_public.yml',
    ]
    for filepath in private_artifacts:
        if not os.path.isfile(filepath):
            raise FileNotFoundError(f"Private delivery artifact expected by public generator is missing: {filepath}")
        os.remove(filepath)
        print(f"✓ Removed private delivery artifact: {filepath}")


def validate_public_go_sources(root='.'):
    """Fail closed when a new private Go import was missed by the generator."""
    private_import = re.compile(
        r'"github\.com/oneclickvirt/(?:privatespeedtest|security)(?:/|\")',
        flags=re.IGNORECASE,
    )
    matches = []
    ignored_directories = {'.git', 'vendor', '.cache', '.tmp'}
    for directory, directories, filenames in os.walk(root):
        directories[:] = [name for name in directories if name not in ignored_directories]
        for filename in filenames:
            if not filename.endswith('.go'):
                continue
            filepath = os.path.join(directory, filename)
            if private_import.search(read_file(filepath)):
                matches.append(filepath)
    if matches:
        raise ValueError(f"Public Go source still imports restricted modules: {', '.join(matches)}")


def validate_public_delivery_tree(root='.'):
    """Reject private-module use from source and executable build inputs only."""
    private_reference = re.compile(
        r'github\.com/oneclickvirt/(?:privatespeedtest|security)(?:/|["`]|$)',
        flags=re.IGNORECASE,
    )
    # Documentation and ordinary text may describe historical components. They
    # are not Go dependencies and must remain intact in the public branch.
    text_extensions = {'.go', '.mod', '.yaml', '.yml'}
    ignored_directories = {'.git', 'vendor', '.cache', '.tmp'}
    matches = []
    for directory, directories, filenames in os.walk(root):
        directories[:] = [name for name in directories if name not in ignored_directories]
        for filename in filenames:
            if os.path.splitext(filename)[1].lower() not in text_extensions:
                continue
            filepath = os.path.join(directory, filename)
            if private_reference.search(read_file(filepath)):
                matches.append(filepath)
    if matches:
        raise ValueError(f"Public delivery tree still references restricted modules: {', '.join(matches)}")


def modify_utils_go(filepath):
    """
    Modify utils/utils.go to:
    1. Replace security/network import with basics/network
    2. Preserve the existing upload-token behavior without that dependency
    """
    content = read_file(filepath)
    
    # basics/network is already imported as bnetwork. Remove the private
    # import rather than importing the same package twice under two aliases.
    content, import_replacements = re.subn(
        r'(?m)^\s*"github\.com/oneclickvirt/security/network"\r?\n',
        '',
        content,
    )
    if import_replacements != 1:
        raise ValueError(f"Expected one private network import in {filepath}")

    content, check_replacements = re.subn(
        r'(?<![A-Za-z0-9_])network\.NetworkCheck',
        'bnetwork.NetworkCheck',
        content,
    )
    if check_replacements != 1:
        raise ValueError(f"Expected one private network check in {filepath}")
    
    # Replace token usage - find the exact line and replace it
    content = re.sub(
        r'\ttoken := network\.SecurityUploadToken',
        r'\ttoken := "OvwKx5qgJtf7PZgCKbtyojSU.MTcwMTUxNzY1MTgwMw"',
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
    activate_public_component('api/components_public.go')
    remove_private_go_sources()
    modify_utils_go('utils/utils.go')
    print()
    
    # Modify go.mod
    print("Modifying go.mod...")
    modify_go_mod('go.mod')
    remove_vendor_tree()
    print()

    print("Removing private-module delivery inputs...")
    remove_private_delivery_artifacts()
    validate_public_go_sources()
    validate_public_delivery_tree()
    print()
        
    print("✓ All modifications completed successfully!")
    print()
    print("Next steps:")
    print("1. Run 'go mod tidy' to clean up dependencies")
    print("2. Run 'go build -o maintest' to verify compilation")
    print("3. Test the binary with: ./maintest -menu=false -l en -security=false -upload=false")


if __name__ == '__main__':
    main()
