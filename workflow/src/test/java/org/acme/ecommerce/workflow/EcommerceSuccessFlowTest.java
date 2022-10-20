package org.acme.ecommerce.workflow;

import org.junit.jupiter.api.Test;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.github.tomakehurst.wiremock.WireMockServer;

import io.quarkus.test.common.QuarkusTestResource;
import io.quarkus.test.junit.QuarkusTest;
import io.restassured.RestAssured;
import io.restassured.http.ContentType;

import static com.github.tomakehurst.wiremock.client.WireMock.postRequestedFor;
import static com.github.tomakehurst.wiremock.client.WireMock.urlEqualTo;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;

@QuarkusTest
@QuarkusTestResource(SuccessRestFunctionsMock.class)
public class EcommerceSuccessFlowTest {

    static {
        RestAssured.enableLoggingOfRequestAndResponseIfValidationFails();
    }

    // Injected by QuarkusTestResource
    WireMockServer successMockServer;

    @Test
    void verifyFlowRunSuccess() throws JsonProcessingException {
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
        assertFalse(response.getWorkflowData().isError());

        // we called all of our 4 functions
        successMockServer.verify(4, postRequestedFor(urlEqualTo("/")));
    }

}
