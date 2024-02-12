package com.egoodhall.nrpc.gen;

import com.google.inject.Guice;
import com.google.inject.Key;
import com.google.inject.TypeLiteral;
import com.google.protobuf.compiler.PluginProtos;
import com.google.protobuf.compiler.PluginProtos.CodeGeneratorResponse.Feature;
import com.google.protobuf.compiler.PluginProtos.CodeGeneratorResponse.File;
import com.salesforce.jprotoc.Generator;
import com.salesforce.jprotoc.GeneratorException;
import com.salesforce.jprotoc.ProtocPlugin;
import java.util.List;

public class NrpcGenerator extends Generator {

  private static final Key<List<File>> FILES_KEY = Key.get(new TypeLiteral<>() {});

  public static void main(String[] args) {
    if (args.length == 0) {
      ProtocPlugin.generate(new NrpcGenerator());
    } else {
      ProtocPlugin.debug(new NrpcGenerator(), args[0]);
    }
  }

  @Override
  protected List<Feature> supportedFeatures() {
    return List.of(Feature.FEATURE_PROTO3_OPTIONAL);
  }

  @Override
  public List<File> generateFiles(PluginProtos.CodeGeneratorRequest request)
    throws GeneratorException {
    return Guice.createInjector(new NrpcModule(request)).getInstance(FILES_KEY);
  }
}
