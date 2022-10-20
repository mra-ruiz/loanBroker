package org.acme.ecommerce.workflow;

import com.fasterxml.jackson.annotation.JsonProperty;

public class WorkflowResponseModel {

    private String id;
    private WorkflowData workflowData;

    public String getId() {
        return id;
    }

    @JsonProperty("workflowdata")
    public WorkflowData getWorkflowData() {
        return workflowData;
    }

    public static final class WorkflowData {
        private boolean error;

        public boolean isError() {
            return error;
        }

        public void setError(boolean error) {
            this.error = error;
        }
    }

}
