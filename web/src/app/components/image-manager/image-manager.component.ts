import { Component, EventEmitter, Input, Output } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../api.service';
import { HttpEventType, HttpResponse } from '@angular/common/http';

@Component({
  selector: 'app-image-manager',
  templateUrl: './image-manager.component.html',
  styleUrls: ['./image-manager.component.scss'],
  standalone: false
})
export class ImageManagerComponent {
  @Input() images: any[] = [];
  @Output() imagesChange = new EventEmitter<any[]>();

  constructor(private apiService: ApiService) {

  }


  ngOnInit(): void {
    this.apiService.getImages().subscribe((images: any) => {
      this.images = images;
    });
  }
  //file upload
  hash: string;
  description: string;
  currentFile?: File;
  progress = 0;
  message = '';
  selectedFiles?: FileList;
  fileInfos?: Observable<any>;

  selectFile(event: Event) {
    const input = event.target as HTMLInputElement;
    this.selectedFiles = input.files;
  }

  upload(): void {
    this.progress = 0;
    if (this.selectedFiles) {
      const file: File | null = this.selectedFiles.item(0);

      if (file) {
        this.currentFile = file;

        this.apiService.addImage(this.currentFile, this.hash, this.description).subscribe(
          (event: any) => {
            if (event.type === HttpEventType.UploadProgress) {
              this.progress = Math.round(100 * event.loaded / event.total);
            } else if (event instanceof HttpResponse) {
              this.message = event.body.message;
              this.images.push(event.body);
            }
          },
          (err: any) => {
            this.progress = 0;

            this.message = err?.error?.message || err?.error?.error_message || 'Could not upload the file!';

            this.currentFile = undefined;
          });
      }

      this.selectedFiles = undefined;
    }
  }

  remove(id: number) {
    this.images = this.images.filter(img => img.id !== id);
    this.imagesChange.emit(this.images);
  }
}
