import 'package:cryptography/cryptography.dart';
import 'package:http/http.dart' as http;
import "package:msgpack_dart/msgpack_dart.dart" as mp;
import 'dart:typed_data';

final aesCtr = AesCtr.with128bits(
  macAlgorithm: Hmac.sha256(),
);

class LinkClient {
  final http.Client _client;
  final String _host;
  final int _port;

  LinkClient(String host, int port)
      : _client = http.Client(),
        _host = host,
        _port = port;

  Future<Map> send(Map msg) async {
    final secretKey = SecretKeyData(
      Uint8List.fromList([1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1]),
      overwriteWhenDestroyed: true,
    );

    var data = mp.serialize(msg);

    final nonce = aesCtr.newNonce();
    final secretBox = await aesCtr.encrypt(
      data,
      secretKey: secretKey,
      nonce: nonce,
    );

    var httpResp = await _client.post(
      Uri(scheme: "https", host: _host, port: _port, path: '/v1'),
      body: secretBox.concatenation(),
    );

    // TODO: Check status code

    final respBox = SecretBox.fromConcatenation(
      httpResp.bodyBytes,
      nonceLength: aesCtr.nonceLength,
      macLength: aesCtr.macAlgorithm.macLength,
    );
    final deResp = await aesCtr.decrypt(respBox, secretKey: secretKey);

    Map deMsg = mp.deserialize(Uint8List.fromList(deResp));
    return deMsg;
  }
}
