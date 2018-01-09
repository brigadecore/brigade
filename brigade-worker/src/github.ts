import {BrigadeEvent} from "./events";

/**
 * Add some common CI environment variables guessed from the GitHub's webhook payload.
 */
export function extractEnv(event: BrigadeEvent): { [key: string]: string } {
    const payload = JSON.parse(event.payload);

    return {
        GITHUB_EVENT_TYPE: event.type,
        GITHUB_PULL_REQUEST: payload.pull_request ? payload.pull_request.number.toString() : "",
        GITHUB_REF: payload.ref ? payload.ref : "",
        GITHUB_REF_TYPE: payload.ref_type ? payload.ref_type : "",
        GITHUB_BASE_REF: payload.pull_request ? payload.pull_request.base.ref : "",
        GITHUB_HEAD_REF: payload.pull_request ? payload.pull_request.head.ref : "",
    };
}
