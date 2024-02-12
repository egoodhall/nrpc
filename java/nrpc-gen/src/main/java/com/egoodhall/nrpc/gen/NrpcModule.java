package com.egoodhall.nrpc.gen;

import com.egoodhall.nrpc.gen.generators.ClientGenerator;
import com.egoodhall.nrpc.gen.generators.FileGenerator;
import com.egoodhall.nrpc.gen.generators.ServerGenerator;
import com.egoodhall.nrpc.gen.generators.ServiceGenerator;
import com.egoodhall.nrpc.gen.providers.GeneratedFilesProvider;
import com.egoodhall.nrpc.gen.providers.ServiceInfoProvider;
import com.google.inject.Binder;
import com.google.inject.Module;
import com.google.inject.Provider;
import com.google.inject.Provides;
import com.google.inject.TypeLiteral;
import com.google.inject.multibindings.Multibinder;
import com.google.protobuf.compiler.PluginProtos;
import com.salesforce.jprotoc.ProtoTypeMap;

class NrpcModule implements Module {

  private final PluginProtos.CodeGeneratorRequest codeGeneratorRequest;

  public NrpcModule(PluginProtos.CodeGeneratorRequest codeGeneratorRequest) {
    this.codeGeneratorRequest = codeGeneratorRequest;
  }

  @Override
  public void configure(Binder binder) {
    bindProvider(binder, new TypeLiteral<>() {}, GeneratedFilesProvider.class);
    bindProvider(binder, new TypeLiteral<>() {}, ServiceInfoProvider.class);
    bindGenerator(binder, ServiceGenerator.class);
    bindGenerator(binder, ServerGenerator.class);
    bindGenerator(binder, ClientGenerator.class);
  }

  private <T> void bindProvider(
    Binder binder,
    TypeLiteral<T> type,
    Class<? extends Provider<T>> provider
  ) {
    binder.bind(type).toProvider(provider);
  }

  private void bindGenerator(Binder binder, Class<? extends FileGenerator> clazz) {
    Multibinder.newSetBinder(binder, FileGenerator.class).addBinding().to(clazz);
  }

  @Provides
  PluginProtos.CodeGeneratorRequest providesCodeGeneratorRequest() {
    return codeGeneratorRequest;
  }

  @Provides
  ProtoTypeMap providesProtoTypeMap(PluginProtos.CodeGeneratorRequest request) {
    return ProtoTypeMap.of(request.getProtoFileList());
  }
}
