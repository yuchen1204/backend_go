import 'package:flutter_test/flutter_test.dart';

import 'package:fultter_package/fultter_package.dart';

void main() {
  test('BackendClient tokens set/get/clear', () {
    final client = BackendClient(baseUrl: 'http://localhost:8080');
    // default empty
    expect(client.getTokens()['accessToken'], isNull);
    expect(client.getTokens()['refreshToken'], isNull);

    client.setTokens(accessToken: 'a', refreshToken: 'r');
    expect(client.getTokens()['accessToken'], 'a');
    expect(client.getTokens()['refreshToken'], 'r');

    client.clearTokens();
    expect(client.getTokens()['accessToken'], isNull);
    expect(client.getTokens()['refreshToken'], isNull);
  });
}
