import {Component} from '@angular/core';
import {SpinnerComponent} from '../spinner'
@Component({
    selector: 'blockui',
    styleUrls: ['./blockui.css'],
    template:
    `<div class="in modal-backdrop block-overlay"></div>
     <div class="block-message-container" aria-live="assertive" aria-atomic="true">
        <div class="block-message" >    <my-spinner style="margin-right: 20px" [isRunning]=true></my-spinner>
            {{ message }}
    </div>
    </div>`
})
export class BlockUIComponent {
    message: any = 'Reloading Config...';

}
