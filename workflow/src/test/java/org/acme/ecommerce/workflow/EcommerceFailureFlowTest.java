package org.acme.ecommerce.workflow;

import org.junit.jupiter.api.Test;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.github.tomakehurst.wiremock.WireMockServer;

import io.quarkus.test.common.QuarkusTestResource;
import io.quarkus.test.junit.QuarkusTest;
import io.restassured.RestAssured;
import io.restassured.http.ContentType;

import static com.github.tomakehurst.wiremock.client.WireMock.containing;
import static com.github.tomakehurst.wiremock.client.WireMock.postRequestedFor;
import static com.github.tomakehurst.wiremock.client.WireMock.urlEqualTo;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.wildfly.common.Assert.assertTrue;

@QuarkusTest
@QuarkusTestResource(FailureRestFunctionsMock.class)
public class EcommerceFailureFlowTest {

    static {
        RestAssured.enableLoggingOfRequestAndResponseIfValidationFails();
    }

    // Injected by QuarkusTestResource
    WireMockServer failureMockServer;

    @Test
    void verifyFlowRunFailure() throws JsonProcessingException {
        final ObjectMapper mapper = new ObjectMapper();

        final String workflowResponse = RestAssured.given()
                .accept(ContentType.JSON)
                .contentType(ContentType.JSON)
                .body("{ \"workflowdata\": { \"error\": false } }")
                .post("/commerce")
                .then()
                .statusCode(201).extract().body().asPrettyString();
        assertNotNull(workflowResponse);

        final WorkflowResponseModel response = mapper.readValue(workflowResponse, WorkflowResponseModel.class);
        assertNotNull(response.getId());
        assertTrue(response.getWorkflowData().isError());

        // first function returned 500, transition to Failure which calls `failure` function, so two times.
        failureMockServer.verify(2, postRequestedFor(urlEqualTo("/")));
        failureMockServer.verify(1, postRequestedFor(urlEqualTo("/")).withRequestBody(containing("true")));
    }

}
