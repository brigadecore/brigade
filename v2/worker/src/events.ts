import { Event, EventRegistry as BrigadierEventRegistry } from "../../brigadier/src/events"

class EventRegistry extends BrigadierEventRegistry {

  public fire(event: Event): void {
    const handlerFn = this.handlers.get(`${event.source}:${event.type}`)
    if (handlerFn) {
      handlerFn(event) 
    }
  }

}

export const events = new EventRegistry()
