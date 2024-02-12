package com.egoodhall.nrpc.example;

import static org.assertj.core.api.Assertions.assertThat;

import com.egoodhall.nrpc.ClientOptions;
import com.egoodhall.nrpc.NrpcError;
import com.egoodhall.nrpc.Result;
import com.egoodhall.nrpc.ServerOptions;
import io.nats.client.Connection;
import io.nats.client.Nats;
import java.io.IOException;
import java.util.function.Consumer;
import nats.io.NatsServerRunner;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

public class EchoServiceTest {

  private static NatsServerRunner nats;

  @BeforeAll
  static void setUp() throws IOException {
    nats = new NatsServerRunner(true);
  }

  @AfterAll
  static void tearDown() throws Exception {
    nats.close();
  }

  @Test
  void itRunsSuccessfully() throws IOException, InterruptedException {
    runTest(client -> {
      Result<EchoReply, NrpcError> echo = client.echo(
        EchoRequest.newBuilder().setMessage("hello world!").build()
      );

      assertThat(echo)
        .matches(Result::isOk, "is Ok")
        .extracting(Result::getOkOrThrow)
        .extracting(EchoReply::getMessage)
        .isEqualTo("hello world!");
    });
  }

  @Test
  void itCanUseNamespaces() throws IOException, InterruptedException {
    ServerOptions serverOptions = ServerOptions.defaults().withNamespace("test");
    ClientOptions clientOptions = ClientOptions.defaults().withNamespace("test");

    runTest(
      serverOptions,
      clientOptions,
      client -> {
        Result<EchoReply, NrpcError> echo = client.echo(
          EchoRequest.newBuilder().setMessage("hello world!").build()
        );

        assertThat(echo)
          .matches(Result::isOk, "is Ok")
          .extracting(Result::getOkOrThrow)
          .extracting(EchoReply::getMessage)
          .isEqualTo("hello world!");
      }
    );
  }

  @Test
  void itCanUseHashNamespaces() throws IOException, InterruptedException {
    ServerOptions serverOptions = ServerOptions
      .builder()
      .setHashNamespace("test")
      .build();
    ClientOptions clientOptions = ClientOptions
      .builder()
      .setHashNamespace("test")
      .build();

    runTest(
      serverOptions,
      clientOptions,
      client -> {
        Result<EchoReply, NrpcError> echo = client.echo(
          EchoRequest.newBuilder().setMessage("hello world!").build()
        );

        assertThat(echo)
          .matches(Result::isOk, "is Ok")
          .extracting(Result::getOkOrThrow)
          .extracting(EchoReply::getMessage)
          .isEqualTo("hello world!");
      }
    );
  }

  private void runTest(Consumer<EchoServiceClient> test)
    throws IOException, InterruptedException {
    runTest(ServerOptions.defaults(), ClientOptions.defaults(), test);
  }

  private void runTest(
    ServerOptions serverOptions,
    ClientOptions clientOptions,
    Consumer<EchoServiceClient> test
  ) throws IOException, InterruptedException {
    EchoService service = new EchoServiceImpl();

    try (Connection connection = Nats.connect(nats.getURI())) {
      try (
        EchoServiceServer ignored = new EchoServiceServer(
          connection,
          service,
          serverOptions
        )
      ) {
        EchoServiceClient client = new EchoServiceClient(connection, clientOptions);

        test.accept(client);
      }
    }
  }
}
