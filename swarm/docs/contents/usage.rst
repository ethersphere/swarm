*****************
Usage
*****************

Using a local Swarm Instance
================================

Running your client
------------------------------
To start a swarm node we must start geth with an empty data directory on a private network. First set aside an empty temporary directory to be the data store

.. code-block:: js

   DATADIR=/tmp/BZZ/`date +%s`

then make a new account using this directory

.. code-block:: js

 PASSWORD="mypassword"
 ./geth --datadir $DATADIR --password  `echo -n $PASSWORD` account new
 BZZKEY=0x1234567890abcdef1234567890abcdef12345678

and finally, launch geth on a private network (id 322)

.. code-block:: js

    ./geth --datadir $DATADIR \
           --bzzaccount $BZZKEY
           --port 30301 \
           --unlock primary \
           --password `echo $PASSWORD` \
           --verbosity 6 \
           --rpc \
           --rpcport 8101 \
           --rpccorsdomain '*' \
           --bzz \
           --networkid 322 \
           --nodiscover \
           --maxpeers 0 \
           console   2>> $DATADIR/bzz.log

At this verbosity level you should see plenty of output accumulating in the logfile. You can keep en eye on it using the command @command{tail -f $DATADIR/bzz.glog}.


Uploading a file or directory to your local swarm instance
---------------------------------------------------------------

Included in the swarm repository is a shell script that makes it easy to upload a file to a local swarm node using http port 8500.

.. code-block:: js

   bash bzz/bzzup/bzzup.sh /path/to/myFileOrDirectory

If this command is successful, the output will be a hash

.. code-block:: js

   65b2a32ab2230d7d2bad2616e804d374921be68758009491cd52c727e37b4979

If unsuccessful (for example if no local node is running) the output will simply be blank.

It is also possible to upload a file or directory from the console like this

.. code-block:: js

    hash = bzz.upload("/path/to/myFileOrDirectory", "index.html")

Here the second parameter (index.html) is to be mapped to the root path '/'.

Downloading a file from your local swarm instance
---------------------------------------------------------

Your local swarm instance has an http interface running on port 8500 (by default). To download a file is thus a simple matter of pointing your browser to

.. code-block:: js

    http://localhost:8500/65b2a32ab2.. .7b4979

or, if you prefer, you can use the console

.. code-block:: js

    bzz.get(hash)


Manifests
================

In general Manifests declare a list of strings associated with swarm entries. Before we get into generalities however, let us begin with an introductory example.

A Manifest example - directory trees
---------------------------------------

Suppose we had used @command{bzzup.sh} (as described above) to upload a directory to swarm instead of just a file:

.. code-block:: js

    bash bzz/bzzup/bzzup.sh /path/to/directory

then the resulting hash points to a "manifest" - in this case a list of files within the directory along with their swarm hashes. Let us take a closer look.

The raw Manifest
-----------------------
We can see the raw Manifest by prepending @code{raw/} to the URL like so

.. code-block:: js

    wget -O - "http://localhost:8500/raw/HASH"

In our example it contains a list of all files contained in @code{/path/to/directory} together with their swarm ids (hashes) as well as their content-types. It may look like this: (whitespace added here to make it legible)

.. code-block:: js

  {"entries":[{"hash":"HASH-for-fileA1",
  "path":"directoryA/fileA1",
  "contentType":"text/plain"},
  {"hash":"HASH-for-fileB2",
  "path":"directoryA/directoryB/fileB2",
  "contentType":"text/plain"},
  {"hash":"HASH-for-fileB1",
  "path":"directoryA/directoryB/fileB1",
  "contentType":"text/plain"},
  {"hash":"HASH-for-fileC1",
  "path":"directoryA/directoryC/fileC1",
  "contentType":"text/plain"}]}


A note on content type
----------------------------


Manifests contain content-type information for the hashes they reference. In other contexts, where content-type is not supplied or, when you suspect the information is wrong, it is possible in a raw query to specify the content-type manually in the search query.

.. code-block:: js

   http.get("http://localhost:8500/raw/hash/?content_type=\"text/plain\"")

Path Matching on Manifests
---------------------------------

A useful feature of manifests is that Urls can be matched on the paths. In some sense this makes the manifest a routing table and so the manifest swarm entry acts as if it were a host.

More concretely, continuing in our example, we can access the file

.. code-block:: js

    /path/to/directory/subdirectory/filename

by pointing the browser to

.. code-block:: js

    http://localhost:8500/HASH/subdirectory/filename

.. note:: if the filename is @code{index.html} then it can be omitted.

Manifests in general
--------------------------

Although in our example above the manifest was essentially a file listing in a directory, there is no reason for a Manifest to take this form. Manifests simply match strings with swarm id's, and there is no requirement that the strings be of the form @code{path/to/file}. Indeed swarm treats @code{path/to/file} as just another identifying string and there is nothing special about the @code{/} character.

@strong{However}, a browser will treat @code{/} as a special character. This is important to remember when specifying (relative) URL's in your Dapp.

The bzz:// URL scheme
========================
To make it easier to access swarm content, we can use the bzz URL scheme. One of its primary merits is that it allows us to use human readable addresses instead of hashes. This is achieved by a name registration contract on the blockchain.

http module for urls on the console
----------------------------------------
The in-console http client understands the bzz scheme if geth is started with swarm enabled. Syntax:

.. code-block:: js

    http.get(url)
    http.download(url, /path/to/save)

The console http module is a very simple http client, that understands the bzz scheme if bzz is enabled.

* `http.get(url)`
* `http.download(url, /path/to/save)`
* `http.loadScript(url)` should be same as JSRE.loadScript

bzz console api overview
----------------------------

  bzz.upload(localfspath, indexfile)
  returns content hash

  bzz.download(bzzpath, localdirpath)

  bzz.put(content, contentType)

   returns content hash

  bzz.get(bzzpath)
  returns object with content, mime type, status code and content size

  bzz.register(address, hash, domain)

  bzz.resolve(domain)
  returns content hash

Name Registration for swarm content
-----------------------------------------

It is the swarm hash of a piece of data that dictates routing. Therefore its role is somehwhat analogous to an IP address in the TCP/IP internet. Domain names can be registered on the blockchain and set to resolve to any swarm hash. The bzz blockchain registry is thus analogous to DNS (and no ICANN nor any name servers are needed).

Currently the domain name is any arbitrary string in that the contract does not impose any restrictions. Since this is used in the host part of the url in the bzz scheme, we recommend using wellformed domain names so that there is interoperability with restrictive url handler libs.

In the bzz:// URL scheme it is possible to supply a block number;

.. code-block:: js

  bzz://swarm.com;144

and this means that we want swarm.com to be resolved to a hash as registered in the registry at block 144. (Note the semicolon @code{;} in the URL)

Example: using bzz api and registered names:

.. code-block:: js

   hash = bzz.upload("/path/to/my/directory");

   hash = bzz.put("console.log(\"hello from console\")", "application/javascript");

  bzz.get(hash);
  {
    content: 'console.log("hello");',
    contentType: 'application/javascript',
    status: '0'
    size: '21',
  }

  http.get("bzz://"+hash);
  'console.log("hello from console")'

  http.loadScript("bzz://"+hash);
  hello from console
  true

  bzz.register(primary, hash, "hello")

Name registration for contracts
-----------------------------------------

It is also possible to register human readable names for contracts.
@subsubheading Prerequisites
In order to do this, you must have a @code{globalRegistrar} contract deployed and you must have HashReg, @code{UrlHint} deployed and registered with @code{globalRegistrar}.

These need to be done only once for every chain. See appendix.

If this was successful, you will see these commands respond with addresses.

.. code-block:: js

  registrar.owner("HashReg");
  registrar.owner("UrlHint");
  registrar.addr("HashReg");
  registrar.addr("UrlHint");


and these commands will respond with code:

.. code-block:: js

  eth.getCode(globalRegistrarAddr);
  eth.getCode(hashRegAddr);
  eth.getCode(urlHintAddr);


If these checks are ok, you are all set up.

Creating and a contract
++++++++++++++++++++++++++++++++

In order to continue this example, we must write a contract and deploy its compiled code on the blockchain. We proceed:

.. code-block:: js

  source = "contract test \n" +
  "   /// @@notice will multiply `a` by 7.\n" +
  "   function multiply(uint a) returns(uint d) {\n" +
  "      return a * 7;\n" +
  "   }\n" +
  "} ";
  contract = eth.compile.solidity(source).test;
  contractaddress = eth.sendTransaction({from: primary, data: contract.code});


Then we must wait until the contract is included in a block. Thus, if we are on a private test network, wem must mine a block

.. code-block:: js

    miner.start(1); admin.sleepBlocks(1); miner.stop();


we continue

.. code-block:: js

  contractaddress = eth.getTransactionReceipt(txhash).contractAddress;
  eth.getCode(contractaddress);

  multiply7 = eth.contract(contract.info.abiDefinition).at(contractaddress);
  fortytwo = multiply7.multiply.call(6);


Then we check if everything worked and the contracts are deployed and usable

.. code-block:: js

  code = eth.getCode(contractaddress);
  abiDef = contract.info.abiDefinition;
  multiply7 = eth.contract(abiDef).at(contractaddress);
  multiply7.multiply.call(6);

Deploying contract info in swarm and registering its hash
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

The contract.info substructure given back from the solidity compiler can be deployed with swarm. The resulting contenthash is registered in the HashReg.


.. code-block:: js

  contenthash = bzz.put(JSON.stringify(contract.info),  "application/eth-contractinfo+json");
  admin.register(primary, contractaddress, contenthash);
  miner.start(1); admin.sleepBlocks(1); miner.stop();
  //mining only needed if you are on a private chain self mining


Contract usage from dapp (or user-side case example)
--------------------------------------------------------------

:command:`eth.getContractInfo()` will magically work. If the url fetcher has the bzz protocol scheme enabled, then it tries to fetch it with the registered contenthash. (If there is no swarm or the content is not (yet) uploaded there, it gracefully falls back to the UrlHint, ie., it looks up the url hint for the contentHash, fetches its content, and verifies it against the contentHash for protection.)

Note that the user needs the contractaddress but nothing else.


.. code-block:: js

  info = admin.getContractInfo(contractaddress);
  multiply7 = eth.contract(info.abiDefinition).at(contractaddress);

Now that we switch on confirmations and try:


.. code-block:: js

  eth.confirmTransactions(true);
  multiply7.multiply.sendTransaction(6, { from: primary });


The following custom confirmation message should appear on the console and 6 shall be multiplied by seven:


.. code-block:: js

  myMultiply7.multiply.sendTransaction(6);
  NatSpec: Will multiply 6 by 7.
  Confirm? [y/n] y


Registering names for contracts
++++++++++++++++++++++++++++++++++++++++

And now we can go one step further and use the globalRegistrar name registry for contracts:


.. code-block:: js

  eth.confirmTransactions(true);
  registrar.reserve.sendTransaction("multiply7", {from:primary});
  registrar.setAddress.sendTransaction("multiply7", contractaddress, true, {from:primary});


You need to wait for these 2 transactions to be confirmed.

.. code-block:: js

  miner.start(1); admin.sleepBlocks(2); miner.stop();

You can check if they arrived:

.. code-block:: js

  registrar.owner("multiply7");

Now the contract name is sufficient to use this contract from a Dapp.

.. code-block:: js

  contractaddress = registrar.addr("multiply7");
  info = admin.getContractInfo(contractaddress);
  multiply7 = eth.contract(info.abiDefinition).at(contractaddress);


If info is only needed because of the Abi, then one could define this function:


.. code-block:: js

  getContract = function(name) {
    contractaddress = registrar.addr(name);
    info = admin.getContractInfo(contractaddress);
    return eth.contract(info.abiDefinition).at(contractaddress);
  }


.. code-block:: js

  web3.sha3(eth.getCode(registrar.addr("multiply7")))
  51b68b0f44e8c6ef096797efbed04185fd4c4a639cd5ffe52e96076519c1385d

Using bzz domain names
-------------------------

Now that we know how to register names, let us see how to use them in practice

.. code-block:: js

  albumHash = bzz.upload("/Users/tron/Work/ethereum/go-ethereum/bzz/bzzdemo/",   "index.html")
  bzz.register(primary, "album", albumHash)
  miner.start(1); admin.sleepBlocks(1); miner.stop();
  //mining needed if you are on a private chain
  bzz.resolve("album")
  admin.httpGet("bzz:/album/")


you can also try

.. code-block:: js

  bzz.download("/album", "/tmp/album");
  bzz.upload("/tmp/album", "index.html");


And using the bzz URL's in the http module we can now try these (matching, fallbacks errors)

.. code-block:: js

  http.get("bzz://51b68b0f44e8c6ef.. .1385d/")
  http.get("bzz://album/index.html")
  http.get("bzz://album/index.css")


As indicated above, we can force a content type manually to get at the raw content:


.. code-block:: js

  http.get("http://raw/album/?content\_type=\"text/plain\"")

Changing registered name, managing versions, rollback
-------------------------------------------------------------

Suppose we have registered the name @code{swarmpicture} as in


.. code-block:: js

  bzz.register(primary, "swarmpicture",     bzz.upload("bzz.demo/swarm-inside.png", "swarm-inside.png"))


After some blocks are mined, this content will become accessible at

.. code-block:: js

   http://localhost:8500/swarmpicture/

and the resolver should work too as:

.. code-block:: js

  bzz.resolve("swarmpicture")
'0x58c604de89bf3ecbbbfc90948b273ae3f956e6106babd5e8bacb3615213d3c2e'


Let us remember this version of "swarmpicture"

.. code-block:: js

  v1 = eth.blockNumber


Now we realise that we have made a mistake and want to include the full logo in our site and se we re-register:

.. code-block:: js

  bzz.register(primary, "swarmpicture",    bzz.upload("bzz.demo/MSTR-Swarm-Logo.jpg", "MSTR-Swarm-Logo.jpg"))


then mine some more @code{miner.start(); admin.sleepBlocks(1); miner.stop();} and then we can resolve as

.. code-block:: js

  bzz.resolve("swarmpicture")
'0x8232b8259393019920d57737c1073c78a6cee18ffa8bfcfdc0cd378a732415a8'

This new registration of "swarmpicture" is stored at a different block

.. code-block:: js

  v2 = eth.blockNumber


The full historical record is addressable:

.. code-block:: js

   http://localhost:8500/swarmpicture;31/
   http://localhost:8500/swarmpicture;32/

And you can see it with the bzz-aware http client:


.. code-block:: js

    http.get("bzz://raw/swarmpicture:"+v1+"?content\_type=text/json") '{"entries":[{"path":"swarm-inside.png","hash":"a41a826e.. .28",  "contentType":"image/png","status":0},{"path":"", "hash":"a41a826e.. .28","contentType":"image/png","status":0}]}'

    http.get("bzz://raw/swarmpicture:"+v2+"?content\_type=text/json") '{"entries":[{"path":"MSTR-Swarm-Logo.jpg","hash":"35e6a17f.. .1d", "contentType":"image/jpeg","status":0},{"path":"", "hash":"35e6a17f.. .1d","contentType":"image/jpeg","status":0}]}'




Appendix - Deploying a Name Registry
------------------------------------------

mine some ether on a private chain
++++++++++++++++++++++++++++++++++++++++

.. code-block:: js

  primary = eth.accounts[0];
  balance = web3.fromWei(eth.getBalance(primary), "ether");

  admin.miner.start(8);
  admin.sleepBlocks(10);
  admin.miner.stop()  ;


mine transactions on a private chain
+++++++++++++++++++++++++++++++++++++++

.. code-block:: js

  eth.getBlockTransactionCount("pending");
  eth.getBlock("pending", true).transactions;

  miner.start(1);
  admin.sleepBlocks(eth.blockNumber+1);
  miner.stop();

  eth.getBlockTransactionCount("pending");


create and deploy GlobalRegistrar, HashReg and UrlHint
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

.. code-block:: js

  primary = eth.accounts[0];
  globalRegistrarAddr = admin.setGlobalRegistrar(primary);
  hashRegAddr = admin.setHashReg(primary);
  urlHintAddr = admin.setUrlHint(primary);


You need to mine or wait till the txs are all picked up.
Initialise the registrar on the new address and check if the other registars are registered:


.. code-block:: js

  registrar = GlobalRegistrar.at(globalRegistrarAddr);
  registrar.owner("HashReg");
  registrar.owner("UrlHint");
  registrar.addr("HashReg");
  registrar.addr("UrlHint");


Next time you only need to specify the address of the GlobalRegistrar (for the live chain it is encoded in the code)


.. code-block:: js

  admin.setGlobalRegistrar("0x6e332ff2d38e8d6f21bee5ab9a1073166382ce33")
  registrar = GlobalRegistrar.at(GlobalRegistrarAddr);
  registrar.owner("HashReg");
  registrar.owner("UrlHint");
  registrar.addr("HashReg");
  registrar.addr("UrlHint");


