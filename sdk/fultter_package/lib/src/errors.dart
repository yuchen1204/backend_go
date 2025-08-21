class BackendApiError implements Exception {
  final int? status;
  final String? code;
  final dynamic payload;
  final String message;

  BackendApiError(this.message, {this.status, this.code, this.payload});

  @override
  String toString() => 'BackendApiError(status: ' 
      '${status ?? '-'}, code: ${code ?? '-'}, message: $message)';
}
