import { Pipe, PipeTransform } from '@angular/core';

@Pipe({name: 'keyParser'})
export class keyParserPipe implements PipeTransform {
  transform(value) : any {
    let keys = [];
    for (let key in value) {
      keys.push({key: key, value: value[key]});
    }
    return keys;
}
}
