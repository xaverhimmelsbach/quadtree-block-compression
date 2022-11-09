# Quadtree Block Compression

## Usage

### Encoding

```sh
go run . -input original.jpg -output encoded.zip
```

### Decoding
```sh
go run . -input encoded.zip -output decoded.jpg
```

### Visualization
Set `Visualization.Enable` to `True` in `config.yml` to generate previews of the quadtree blocks and the encoded picture in the input size and with added padding.
