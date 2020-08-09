import { v4 as uuidv4 } from "uuid";
import * as store from "store";

class User {
    id: string

    constructor() {
        this.id = store.get('anonymous_id')
        if (!this.id) {
            this.id = uuidv4()
            store.set('anonymous_id', this.id);
        }
    }
}

const currentUser = new User();

export default currentUser;