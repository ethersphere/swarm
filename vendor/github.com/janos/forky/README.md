# Forky

Ethereum Swarm Chunk Store experiment. Intended only for demonstration of possible performance improvements.

## Etymology

Name _forky_ comes from the default Swarm fixed chunk data size of 4096 bytes or 4K (four kay -> forky) bytes, as this storage supports only fixed sized data chunks.

It is also a Toy Story 4 character. Given that Debian project uses Toy Story character names for release codenames, such correlation is quite nice.

## Tests

Default tests are configured to validate correctness of implementations:

```
go test -v 
```

To see results with different number of chunks tests can be run with optional arguments:

```
go test -timeout 30m -v github.com/janos/forky/leveldb -chunks 1000000
```

This will run both plain LevelDB store and Forky with LevelDB MetaStore tests with timings for comparison. A high number of chunks require setting an appropriate timeout flag, also.


## License

The forky library is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html), also
included in our repository in the `COPYING` file.
