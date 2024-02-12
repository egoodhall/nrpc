package com.egoodhall.nrpc;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.MessageLite;
import io.nats.client.Message;

public class Results {

  private static final String ERROR_MESSAGE_HEADER = "Nats-Service-Error";
  private static final String ERROR_CODE_HEADER = "Nats-Service-Error-Code";

  private Results() {}

  @FunctionalInterface
  public interface OkParser<T> {
    T parse(byte[] bytes) throws InvalidProtocolBufferException;
  }

  public static <T extends MessageLite> Result<T, NrpcError> fromThrowable(Throwable t) {
    return Result.err(NrpcError.of(500, t.getMessage()));
  }

  public static <T extends MessageLite> Result<T, NrpcError> fromMessage(
    Message message,
    OkParser<T> parser
  ) {
    if (
      message.hasHeaders() &&
      message.getHeaders().containsKeyIgnoreCase(ERROR_CODE_HEADER) &&
      message.getHeaders().containsKey(ERROR_MESSAGE_HEADER)
    ) {
      return Result.err(
        NrpcError.of(
          Integer.parseInt(message.getHeaders().getFirst(ERROR_CODE_HEADER)),
          message.getHeaders().getFirst(ERROR_MESSAGE_HEADER)
        )
      );
    }

    try {
      return Result.ok(parser.parse(message.getData()));
    } catch (InvalidProtocolBufferException e) {
      return Result.err(NrpcError.of(500, e.getMessage()));
    }
  }
}
