# Sinogram

A steganographic encoder that hides binary data within Chinese text by mapping pairs of base64 characters to individual Chinese characters.

## What It Does

Sinogram converts any file into text that looks like a natural Chinese article, but actually encodes your data:

1. **Encodes** your file to base64 (optional)
2. **Maps** pairs of base64 characters (like "AB", "CD") to single Chinese characters (like "科", "技")
3. **Creates** text that appears to be Chinese prose but contains your hidden data

This reduces character count by 50% compared to raw base64, while the UTF-8 byte size increases by ~1.5x.

## Installation

```bash
go build -o sinogram main.go
```

Or run directly:
```bash
go run main.go [flags]
```

## Quick Start

### 1. Generate a Dictionary
First, create a dictionary of Chinese characters:

```bash
./sinogram -gen-dict
```

This creates `dictionary.md` containing Chinese text used for character mapping.

### 2. Encode a File

```bash
./sinogram -e myfile.txt -o encoded.txt
```

Your file is now encoded as Chinese text!

### 3. Decode Back

```bash
./sinogram -d encoded.txt -o restored.txt
```

Your original file is restored.

## Usage

```
./sinogram [flags]

Flags:
  -e string
        Encode: specify input file to encode
  -d string
        Decode: specify input file to decode
  -o string
        Output file name (default: input + .encoded or .decoded)
  -dict string
        Dictionary file path (default: dictionary.md)
  -b64
        Use base64 encoding (default: true)
  -gen-dict
        Generate sample dictionary file
```

## Examples

**Encode an image:**
```bash
./sinogram -e photo.jpg -o photo_encoded.txt
```

**Decode without base64 (for text files):**
```bash
./sinogram -e message.txt -b64=false -o encoded.txt
./sinogram -d encoded.txt -b64=false -o message.txt
```

**Use custom dictionary:**
```bash
./sinogram -e file.pdf -dict my_chinese_text.txt -o output.txt
```

## Limitations

- **Not compression**: The output is actually ~1.5x larger in bytes
- **Requires dictionary**: Both encode and decode need the same dictionary
- **Incomplete coverage**: If dictionary has fewer than 4,096 characters, some pairs remain as base64
- **Not encryption**: This is encoding/steganography, not secure encryption

## Use Cases

- Hide data in plain sight within Chinese text documents
- Create text-based encodings that pass through text-only systems
- Educational purposes for understanding encoding schemes
- Artistic or creative data representation projects