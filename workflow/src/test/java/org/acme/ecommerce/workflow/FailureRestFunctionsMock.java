package org.acme.ecommerce.workflow;

import java.util.HashMap;
import java.util.Map;

import javax.ws.rs.core.MediaType;

import com.github.tomakehurst.wiremock.WireMockServer;
import com.github.tomakehurst.wiremock.core.WireMockConfiguration;

import io.quarkus.test.common.QuarkusTestResourceLifecycleManager;

import static com.github.tomakehurst.wiremock.client.WireMock.aResponse;
import static com.github.tomakehurst.wiremock.client.WireMock.post;
import static com.github.tomakehurst.wiremock.client.WireMock.urlEqualTo;

public class FailureRestFunctionsMock implements QuarkusTestResourceLifecycleManager {

    private WireMockServer wireMockServer;

    @Override
    public Map<String, String> start() {
        wireMockServer = new WireMockServer(WireMockConfiguration.wireMockConfig().dynamicPort());
        wireMockServer.start();
        // just return an empty object, doesn't matter since the workflow does nothing with the data
        wireMockServer.stubFor(post(urlEqualTo("/"))
                .willReturn(aResponse()
                        .withHeader("Content-Type", MediaType.APPLICATION_JSON)
                        .withBody("{}")
                        .withStatus(500)));

        // override the properties in the test environment with the mock server
        Map<String, String> restProperties = new HashMap<>();
        restProperties.put("kogito.sw.functions.orderNew.host", "localhost");
        restProperties.put("kogito.sw.functions.orderNew.port", String.valueOf(wireMockServer.port()));
        restProperties.put("kogito.sw.functions.payment.host", "localhost");
        restProperties.put("kogito.sw.functions.payment.port", String.valueOf(wireMockServer.port()));
        restProperties.put("kogito.sw.functions.inventoryReserve.host", "localhost");
        restProperties.put("kogito.sw.functions.inventoryReserve.port", String.valueOf(wireMockServer.port()));
        restProperties.put("kogito.sw.functions.success.host", "localhost");
        restProperties.put("kogito.sw.functions.success.port", String.valueOf(wireMockServer.port()));
        restProperties.put("kogito.sw.functions.failure.host", "localhost");
        restProperties.put("kogito.sw.functions.failure.port", String.valueOf(wireMockServer.port()));

        return restProperties;
    }

    @Override
    public void stop() {
        if (wireMockServer != null) {
            wireMockServer.stop();
        }
    }

    @Override
    public void inject(Object testInstance) {
        if (testInstance instanceof EcommerceFailureFlowTest) {
            ((EcommerceFailureFlowTest) testInstance).failureMockServer = wireMockServer;
        }
    }
}
